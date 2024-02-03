package stats

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"

	"github.com/dmtaylor/costanza/internal/model"
)

// TODO add interface for this & rename struct. Should have done this from the beginning but whatever
// TODO logging activity is the same function for different tables. Create a more generic method for logging updates
// and pass in table name as argument

type Stats struct {
	pool model.DbPool
}

func New(pool model.DbPool) Stats {
	return Stats{
		pool,
	}
}

func (s Stats) LogActivity(ctx context.Context, guildId, userId uint64, reportMonth string) error {
	var err error

	var existingLogId uint
	err = s.pool.QueryRow(ctx, `
SELECT id
FROM discord_usage_stats
WHERE guild_id = $1 AND user_id = $2 AND report_month = $3
`, guildId, userId, reportMonth).Scan(&existingLogId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No usage so far, insert new record
			_, err = s.pool.Exec(ctx, "INSERT INTO discord_usage_stats(guild_id, user_id, report_month) VALUES ($1, $2, $3)",
				guildId,
				userId,
				reportMonth)
			if err != nil {
				return fmt.Errorf("failed to insert new record: %w", err)
			}
			return nil
		} else {
			// Error case
			return fmt.Errorf("failed to get existing stat: %w", err)
		}
	}
	_, err = s.pool.Exec(ctx, `
UPDATE discord_usage_stats
SET message_count = message_count + 1
WHERE id = $1
`, existingLogId)
	if err != nil {
		return fmt.Errorf("failed to increment message stat: %w", err)
	}

	return nil
}

func (s Stats) LogReaction(ctx context.Context, guildId, userId uint64, reportMonth string) error {

	var existingLogId uint
	err := s.pool.QueryRow(ctx, `
SELECT id
FROM discord_reaction_stats
WHERE guild_id = $1 AND user_id = $2 AND report_month = $3
`, guildId, userId, reportMonth).Scan(&existingLogId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.pool.Exec(ctx, "INSERT INTO discord_reaction_stats(guild_id, user_id, report_month) VALUES ($1, $2, $3)",
				guildId,
				userId,
				reportMonth,
			)
			if err != nil {
				return fmt.Errorf("failed to insert new reaction record: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to get existing reaction stat: %w", err)
		}
	}
	_, err = s.pool.Exec(ctx, `
UPDATE discord_reaction_stats
SET message_count = message_count + 1
WHERE id = $1
`, existingLogId)
	if err != nil {
		return fmt.Errorf("failed to increment reaction stat: %w", err)
	}

	return nil
}

func (s Stats) GetReactionLeadersForMonth(ctx context.Context, guildId uint64, reportMonth string) ([]*model.DiscordReactionScore, error) {
	var scores []*model.DiscordReactionScore
	err := pgxscan.Select(ctx, s.pool, &scores, `
SELECT drs.guild_id, drs.user_id, drs.report_month, drs.message_count - COALESCE(dus.message_count, 0) AS score
FROM discord_reaction_stats drs LEFT OUTER JOIN discord_usage_stats dus USING (guild_id, user_id, report_month)
WHERE drs.guild_id = $1 AND drs.report_month = $2
ORDER BY score DESC
LIMIT 5`, guildId, reportMonth)
	if err != nil {
		return nil, fmt.Errorf("reaction leader query failed: %w", err)
	}

	return scores, nil
}

func (s Stats) RemoveReactionLogForMonth(ctx context.Context, reportMonth string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM discord_reaction_stats WHERE report_month = $1", reportMonth)
	if err != nil {
		return fmt.Errorf("failed to delete reaction stats: %w", err)
	}
	return nil
}

func (s Stats) GetLeaders(ctx context.Context, guildId uint64, reportMonth string) ([]*model.DiscordUsageStat, error) {
	var stats []*model.DiscordUsageStat
	err := pgxscan.Select(ctx, s.pool, &stats, `
SELECT *
FROM discord_usage_stats
WHERE guild_id = $1 AND report_month = $2
ORDER BY message_count DESC
LIMIT 5
`, guildId, reportMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to pull top post stats: %w", err)
	}
	return stats, nil
}

func (s Stats) RemoveMonthActivity(ctx context.Context, reportMonth string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM discord_usage_stats WHERE report_month = $1", reportMonth)
	if err != nil {
		return fmt.Errorf("failed to delete usage stats: %w", err)
	}
	return nil
}

func (s Stats) LogDailyGameActivity(ctx context.Context, gamePlay model.DailyGamePlay, reportMonth string) error {
	var gameWinStat model.DailyGameWinStat
	err := pgxscan.Get(ctx, s.pool, &gameWinStat, `
SELECT *
FROM daily_game_win_stats
WHERE guild_id = $1 AND user_id = $2 AND report_month = $3`, gamePlay.GuildId, gamePlay.UserId, reportMonth)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var winCount int
			var currentStreak int
			var maxStreak int
			if gamePlay.Win {
				winCount = 1
				currentStreak = 1
				maxStreak = 1
			}
			_, err := s.pool.Exec(ctx, `
INSERT INTO daily_game_win_stats(guild_id, user_id, report_month, play_count, guess_count, win_count, current_streak, max_streak)
VALUES ($1, $2, $3, 1, $4, $5, $6, $7)`, gamePlay.GuildId, gamePlay.UserId, reportMonth, gamePlay.Tries, winCount, currentStreak, maxStreak)
			if err != nil {
				return fmt.Errorf("failed to insert new row for game stats: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get existing game stat row: %w", err)
		}
	} else {
		gameWinStat.GuessCount += int(gamePlay.Tries)
		if gamePlay.Win {
			gameWinStat.WinCount += 1
			gameWinStat.CurrentStreak += 1
			if gameWinStat.CurrentStreak > gameWinStat.MaxStreak {
				gameWinStat.MaxStreak = gameWinStat.CurrentStreak
			}
		} else {
			gameWinStat.CurrentStreak = 0
		}
		_, err := s.pool.Exec(ctx, `
UPDATE daily_game_win_stats
SET play_count = play_count + 1, guess_count = $1, win_count = $2, current_streak = $3, max_streak = $4
WHERE id = $5`, gameWinStat.GuessCount, gameWinStat.WinCount, gameWinStat.CurrentStreak, gameWinStat.MaxStreak, gameWinStat.Id)
		if err != nil {
			return fmt.Errorf("failed to update win stats: %w", err)
		}

	}

	return nil
}

func (s Stats) GetDailyGameLeaders(ctx context.Context, guildId uint64, reportMonth string) ([]*model.DailyGameWinStat, error) {
	var gameLeaders []*model.DailyGameWinStat

	err := pgxscan.Select(ctx, s.pool, &gameLeaders, `
SELECT *
FROM daily_game_win_stats
WHERE guild_id = $1 AND report_month = $2
ORDER BY win_count DESC
LIMIT 5`, guildId, reportMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to pull top winners: %w", err)
	}

	return gameLeaders, nil
}

func (s Stats) RemoveDailyGameLeadersForMonth(ctx context.Context, reportMonth string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM daily_game_win_stats WHERE report_month = $1`, reportMonth)
	if err != nil {
		return fmt.Errorf("failed to delete stats for month: %w", err)
	}
	return nil
}

func (s Stats) LogCursedChannelPost(ctx context.Context, guildId, userId uint64, reportMonth string) error {
	var existingLogId uint
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM discord_cursed_channel_stats WHERE guild_id = $1 AND user_id = $2 AND report_month = $3",
		guildId,
		userId,
		reportMonth).Scan(&existingLogId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.pool.Exec(ctx,
				"INSERT INTO discord_cursed_channel_stats(guild_id, user_id, report_month) VALUES ($1, $2, $3)",
				guildId,
				userId,
				reportMonth,
			)
			if err != nil {
				return fmt.Errorf("failed to insert new cursed channel post record: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to get existing cursed channel post record: %w", err)
		}
	}
	_, err = s.pool.Exec(ctx,
		"UPDATE discord_cursed_channel_stats SET message_count = message_count + 1 WHERE id = $1",
		existingLogId)
	if err != nil {
		return fmt.Errorf("failed to increment cursed channel post record: %w", err)
	}
	return nil
}

func (s Stats) RemoveCursedChannelPostStatsForMonth(ctx context.Context, reportMonth string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM discord_cursed_channel_stats WHERE report_month = $1", reportMonth)
	if err != nil {
		return fmt.Errorf("failed to delete cursed channel post stats: %w", err)
	}
	return nil
}

func (s Stats) GetTopCursedChannelPosters(ctx context.Context, guildId uint64, reportMonth string) ([]*model.CursedChannelPost, error) {
	var results []*model.CursedChannelPost
	err := pgxscan.Select(
		ctx,
		s.pool,
		&results,
		"SELECT * FROM discord_cursed_channel_stats WHERE guild_id = $1 AND report_month = $2 ORDER BY message_count DESC LIMIT 5",
		guildId,
		reportMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursed channel post leaders: %w", err)
	}
	return results, nil
}

func (s Stats) LogCursedPost(ctx context.Context, guildId, userId uint64, reportMonth string, count int) error {
	var existingLogId uint
	err := s.pool.QueryRow(ctx,
		"SELECT id FROM discord_cursed_posts_stats WHERE guild_id = $1 AND user_id = $2 AND report_month = $3",
		guildId,
		userId,
		reportMonth).Scan(&existingLogId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.pool.Exec(ctx,
				"INSERT INTO discord_cursed_posts_stats(guild_id, user_id, report_month, message_count) VALUES ($1, $2, $3, $4)",
				guildId,
				userId,
				reportMonth,
				count,
			)
			if err != nil {
				return fmt.Errorf("failed to insert new cursed channel post record: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to get existing cursed channel post record: %w", err)
		}
	}
	_, err = s.pool.Exec(ctx,
		"UPDATE discord_cursed_posts_stats SET message_count = message_count + $1 WHERE id = $2",
		count, existingLogId)
	if err != nil {
		return fmt.Errorf("failed to increment cursed channel post record: %w", err)
	}
	return nil
}

func (s Stats) RemoveCursedPostStatsForMonth(ctx context.Context, reportMonth string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM discord_cursed_posts_stats WHERE report_month = $1", reportMonth)
	if err != nil {
		return fmt.Errorf("failed to delete cursed channel post stats: %w", err)
	}
	return nil
}

func (s Stats) GetTopCursedPosters(ctx context.Context, guildId uint64, reportMonth string) ([]*model.CursedPostStat, error) {
	var results []*model.CursedPostStat
	err := pgxscan.Select(
		ctx,
		s.pool,
		&results,
		"SELECT * FROM discord_cursed_posts_stats WHERE guild_id = $1 AND report_month = $2 ORDER BY message_count DESC LIMIT 5",
		guildId,
		reportMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cursed channel post leaders: %w", err)
	}
	return results, nil
}

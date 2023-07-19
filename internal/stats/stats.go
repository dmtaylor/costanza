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

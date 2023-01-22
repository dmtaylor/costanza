package stats

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"github.com/dmtaylor/costanza/internal/db"
)

type Stats struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Stats {
	return Stats{
		pool,
	}
}

func (s Stats) LogActivity(ctx context.Context, guildId, userId uint64, reportMonth string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get pool connection")
	}
	defer conn.Release()

	var existingLogId uint
	err = conn.QueryRow(ctx, `
SELECT id
FROM discord_usage_stats
WHERE guild_id = $1 AND user_id = $2 AND report_month = $3
`, guildId, userId, reportMonth).Scan(&existingLogId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No usage so far, insert new record
			_, err = conn.Exec(ctx, "INSERT INTO discord_usage_stats(guild_id, user_id, report_month) VALUES ($1, $2, $3)",
				guildId,
				userId,
				reportMonth)
			if err != nil {
				return errors.Wrap(err, "failed to insert new record")
			}
			return nil
		} else {
			// Error case
			return errors.Wrap(err, "failed to get existing stat")
		}
	}
	_, err = conn.Exec(ctx, `
UPDATE discord_usage_stats
SET message_count = message_count + 1
WHERE id = $1
`, existingLogId)
	if err != nil {
		return errors.Wrap(err, "failed to increment message stat")
	}

	return nil
}

func (s Stats) GetLeaders(ctx context.Context, guildId uint64, reportMonth string) ([]*db.DiscordUsageStat, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pool connection")
	}
	defer conn.Release()

	var stats []*db.DiscordUsageStat
	err = pgxscan.Select(ctx, conn, &stats, `
SELECT *
FROM discord_usage_stats
WHERE guild_id = $1 AND report_month = $2
ORDER BY message_count DESC
LIMIT 5
`, guildId, reportMonth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull top post stats")
	}
	return stats, nil
}

func (s Stats) RemoveMonthActivity(ctx context.Context, reportMonth string) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get pool connection")
	}
	defer conn.Release()
	_, err = conn.Exec(ctx, "DELETE FROM discord_usage_stats WHERE report_month = $1", reportMonth)
	if err != nil {
		return errors.Wrap(err, "failed to delete usage stats")
	}
	return nil
}

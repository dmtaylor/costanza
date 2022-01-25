package initiative

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type InitiativeTracker struct {
	pool *pgxpool.Pool
}

type initOrder struct {
}

type initUser struct {
	id            uint64
	initiativeId  uint64
	userSnowflake uint64
	userOrder     uint
}

func NewInitiativeTracker(pool *pgxpool.Pool) InitiativeTracker {
	return InitiativeTracker{
		pool,
	}
}

func (i InitiativeTracker) StartInit(owner uint64, participants []uint64) (uint64, error) {
	// TODO implement logic for creating initiative
	ctx := context.Background()
	conn, err := i.pool.Acquire(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get connection")
	}
	defer conn.Release()

	// Use transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback(ctx)
	query := `INSERT INTO initiative_orders(owner_snowflake, size) VALUES ($1, $2) RETURNING id`
	var init_id uint64
	err = conn.QueryRow(ctx, query, owner, len(participants)).Scan(&init_id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create order")
	}
	users := make([]initUser, len(participants))
	for idx, userSnowflake := range participants {
		users[idx] = initUser{
			initiativeId:  init_id,
			userSnowflake: userSnowflake,
			userOrder:     uint(idx),
		}
	}

	// TODO remove stub return
	return init_id, nil
}

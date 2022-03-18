package initiative

import (
	"context"

	"github.com/jackc/pgx/v4"
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
	err = tx.QueryRow(ctx, query, owner, len(participants)).Scan(&init_id)
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

	insertedCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"initiative_users"},
		[]string{"initiativeId", "user_snowflake", "user_order"},
		pgx.CopyFromSlice(len(users), func(i int) ([]interface{}, error) {
			return []interface{}{users[i].initiativeId, users[i].userSnowflake, users[i].userOrder}, nil
		}),
	)
	if err != nil {
		return init_id, errors.Wrapf(err, "failed to insert users for initiative %d", init_id)
	}
	if insertedCount != int64(len(users)) {
		return init_id, errors.Errorf("mismatched inserted count when adding users for init %d", init_id)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return init_id, errors.Wrap(err, "failed to commit init creation")
	}

	return init_id, nil
}

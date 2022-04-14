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

type InitOrder struct {
	Id             uint64
	OwnerSnowflake uint64
	Size           uint
	CurrentInOrder uint
}

type InitUser struct {
	Id            uint64
	InitiativeId  uint64
	UserSnowflake uint64
	UserOrder     uint
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
	err = tx.QueryRow(ctx, query, owner, len(participants)-1).Scan(&init_id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create order")
	}
	users := make([]InitUser, len(participants))
	for idx, userSnowflake := range participants {
		users[idx] = InitUser{
			InitiativeId:  init_id,
			UserSnowflake: userSnowflake,
			UserOrder:     uint(idx),
		}
	}

	insertedCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"initiative_users"},
		[]string{"initiativeId", "user_snowflake", "user_order"},
		pgx.CopyFromSlice(len(users), func(i int) ([]interface{}, error) {
			return []interface{}{users[i].InitiativeId, users[i].UserSnowflake, users[i].UserOrder}, nil
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

func (i InitiativeTracker) GetUserUp(ctx context.Context, initOrder InitOrder) (InitUser, error) {
	conn, err := i.pool.Acquire(ctx)
	if err != nil {
		return InitUser{}, errors.Wrap(err, "failed to get connection")
	}
	defer conn.Release()
	query := `
	SELECT iu.id, iu.user_snowflake, iu.user_order
	FROM initiative_users iu, initiative_orders io
	WHERE iu.initiative_id = io.id AND iu.user_order = io.current_in_order
	`
	var userId uint64
	var userSnowflake uint64
	var orderPosition uint
	err = conn.Conn().QueryRow(ctx, query).Scan(&userId, &userSnowflake, &orderPosition)
	if err != nil {
		return InitUser{}, errors.Wrap(err, "failed to get current user via query")
	}

	return InitUser{
		userId,
		initOrder.Id,
		userSnowflake,
		orderPosition,
	}, nil
}

/** TODO determine if this is needed
func (i InitiativeTracker) getMostRecentForOwner(ctx context.Context, owner uint64) (initOrder, error) {
	conn, err := i.pool.Acquire(ctx)
	if err != nil {
		return initOrder{}, errors.Wrap(err, "failed to get connection")
	}
	defer conn.Release()

	query := `
	SELECT id, size, current_in_order
	FROM initiative_orders
	WHERE owner_snowflake = $1
	ORDER BY id DESC
	LIMIT 1
	`
	var id uint64
	var size uint
	var current uint
	err = conn.Conn().QueryRow(ctx, query, owner).Scan(&id, &size, &current)
	if err != nil {
		return initOrder{}, errors.Wrap(err, "failed to get most recent order for user")
	}

	return initOrder{
		id,
		owner,
		size,
		current,
	}, nil
}
*/

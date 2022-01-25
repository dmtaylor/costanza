package quotes

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

const GET_QUOTE_QUERY = `
SELECT quote
FROM quotes
WHERE id = $1
`

type QuoteEngine struct {
	rng    *rand.Rand
	lock   sync.Mutex
	dbPool *pgxpool.Pool
	size   uint
}

func NewQuoteEngine(connPool *pgxpool.Pool) (*QuoteEngine, error) {
	ctx := context.Background()
	conn, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get connection")
	}
	defer conn.Release()
	size, err := getQuoteCount(ctx, conn.Conn())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get quote count")
	}
	engine := QuoteEngine{
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:   sync.Mutex{},
		dbPool: connPool,
		size:   size,
	}

	return &engine, nil
}

func (q *QuoteEngine) GetQuoteSql(ctx context.Context) (string, error) {
	q.lock.Lock()
	idx := q.rng.Intn(int(q.size))
	q.lock.Unlock()
	var result string
	err := q.dbPool.QueryRow(ctx, GET_QUOTE_QUERY, idx).Scan(&result)
	if err != nil {
		return "", errors.Wrap(err, "failed to query quote")
	}
	return result, nil
}

func getQuoteCount(ctx context.Context, conn *pgx.Conn) (uint, error) {
	var size uint
	err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM quotes").Scan(&size)
	if err != nil {
		return 0, errors.Wrap(err, "query failed")
	}
	return size, nil
}

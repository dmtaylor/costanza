package quotes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/rand"

	"github.com/dmtaylor/costanza/internal/db"
)

const getQuoteQuery = `
SELECT id, data, type
FROM quotes
WHERE id = $1
`

type QuoteEngine interface {
	GetQuoteSql(ctx context.Context) (db.Quote, error)
}

type QuoteEngineImpl struct {
	rng    *rand.Rand
	lock   sync.Mutex
	dbPool *pgxpool.Pool
	size   uint
}

func NewQuoteEngine(connPool *pgxpool.Pool) (*QuoteEngineImpl, error) {
	ctx := context.Background()
	conn, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Release()
	size, err := getQuoteCount(ctx, conn.Conn())
	if err != nil {
		return nil, fmt.Errorf("failed to get quote count: %w", err)
	}
	src := &rand.PCGSource{}
	src.Seed(uint64(time.Now().UnixNano()))
	engine := &QuoteEngineImpl{
		rng:    rand.New(src),
		lock:   sync.Mutex{},
		dbPool: connPool,
		size:   size,
	}

	return engine, nil
}

func (q *QuoteEngineImpl) GetQuoteSql(ctx context.Context) (db.Quote, error) {
	q.lock.Lock()
	idx := q.rng.Intn(int(q.size))
	q.lock.Unlock()
	conn, err := q.dbPool.Acquire(ctx)
	if err != nil {
		return db.Quote{}, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()
	rows, err := conn.Query(ctx, getQuoteQuery, idx)
	if err != nil {
		return db.Quote{}, fmt.Errorf("failed to execute query: %w", err)
	}
	var result db.Quote
	if err := pgxscan.ScanOne(&result, rows); err != nil {
		return db.Quote{}, fmt.Errorf("failed to scan result: %w", err)
	}
	return result, nil
}

func getQuoteCount(ctx context.Context, conn *pgx.Conn) (uint, error) {
	var size uint
	err := conn.QueryRow(ctx, "SELECT COUNT(*) FROM quotes").Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("query failed: %w", err)
	}
	return size, nil
}

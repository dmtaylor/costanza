package quotes

import (
	"context"
	"fmt"
	"sync"

	"github.com/georgysavva/scany/v2/pgxscan"
	"golang.org/x/exp/rand"

	"github.com/dmtaylor/costanza/internal/model"
)

const getQuoteQuery = `
SELECT id, data, type
FROM quotes
WHERE id = $1
`

type QuoteEngine interface {
	GetQuoteSql(ctx context.Context) (model.Quote, error)
	GetQuoteById(ctx context.Context, id int) (model.Quote, error)
}

type QuoteEngineImpl struct {
	rng    *rand.Rand
	lock   sync.Mutex
	dbPool model.DbPool
	size   uint
}

func NewQuoteEngine(connPool model.DbPool, seed uint64) (*QuoteEngineImpl, error) {
	ctx := context.Background()
	size, err := getQuoteCount(ctx, connPool)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote count: %w", err)
	}
	src := &rand.PCGSource{}
	src.Seed(seed)
	engine := &QuoteEngineImpl{
		rng:    rand.New(src),
		lock:   sync.Mutex{},
		dbPool: connPool,
		size:   size,
	}

	return engine, nil
}

func (q *QuoteEngineImpl) GetQuoteSql(ctx context.Context) (model.Quote, error) {
	q.lock.Lock()
	idx := q.rng.Intn(int(q.size))
	q.lock.Unlock()
	return q.GetQuoteById(ctx, idx)
}

func (q *QuoteEngineImpl) GetQuoteById(ctx context.Context, id int) (model.Quote, error) {
	rows, err := q.dbPool.Query(ctx, getQuoteQuery, id)
	if err != nil {
		return model.Quote{}, fmt.Errorf("failed to execute query: %w", err)
	}
	var result model.Quote
	if err = pgxscan.ScanOne(&result, rows); err != nil {
		return model.Quote{}, fmt.Errorf("failed to scan result: %w", err)
	}
	return result, nil
}

func getQuoteCount(ctx context.Context, db model.DbPool) (uint, error) {
	var size uint
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM quotes").Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("query failed: %w", err)
	}
	return size, nil
}

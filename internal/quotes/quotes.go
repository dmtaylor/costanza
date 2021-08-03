package quotes

import (
	"database/sql"
	_ "embed"
	"math/rand"
	"sync"
	"time"

	"github.com/dmtaylor/costanza/internal/util"
	"github.com/pkg/errors"
)

const GET_QUOTE_QUERY = `
SELECT quote
FROM quotes
WHERE id = ?
`

type QuoteEngine struct {
	rng      *rand.Rand
	lock     sync.Mutex
	stmtPool *util.StatementPool
	size     uint
}

func NewQuoteEngine(connPool *util.SqliteConnectionPool) (*QuoteEngine, error) {
	stmtPool, err := util.NewStatementPool(connPool, GET_QUOTE_QUERY)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stmt pool")
	}
	conn := connPool.Checkout()
	defer connPool.Checkin(conn)
	size, err := getQuoteCount(conn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get quote count")
	}
	engine := QuoteEngine{
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:     sync.Mutex{},
		stmtPool: stmtPool,
		size:     size,
	}

	return &engine, nil
}

func (q *QuoteEngine) GetQuoteSql() (string, error) {
	q.lock.Lock()
	idx := q.rng.Intn(int(q.size))
	q.lock.Unlock()
	stmt := q.stmtPool.Checkout()
	defer q.stmtPool.Checkin(stmt)
	var result string
	err := stmt.QueryRow(idx).Scan(&result)
	if err != nil {
		return "", errors.Wrap(err, "failed to query quote")
	}
	return result, nil
}

func getQuoteCount(conn *sql.DB) (uint, error) {
	var size uint
	err := conn.QueryRow("SELECT COUNT(*) FROM quotes").Scan(&size)
	if err != nil {
		return 0, errors.Wrap(err, "query failed")
	}
	return size, nil
}

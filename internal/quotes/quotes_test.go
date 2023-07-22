package quotes

import (
	"context"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"github.com/dmtaylor/costanza/internal/model"
)

// TODO add more robust tests here, including errors

func TestNewQuoteEngine(t *testing.T) {
	// expected number of rows
	var rowCount uint = 20
	var testSeed uint64 = 2222
	mockdb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	rows := pgxmock.NewRows([]string{"count"}).AddRow(rowCount)
	mockdb.ExpectQuery(`SELECT COUNT\(\*\) FROM quotes`).WillReturnRows(rows)

	testSrc := &rand.PCGSource{}
	testSrc.Seed(testSeed)

	expectedEngine := &QuoteEngineImpl{rng: rand.New(testSrc), lock: sync.Mutex{}, dbPool: mockdb, size: rowCount}
	got, err := NewQuoteEngine(mockdb, testSeed)
	require.Nil(t, err, "got error building quote engine")
	assert.Equal(t, expectedEngine, got, "quote engine doesn't match expected")
	assert.Nil(t, mockdb.ExpectationsWereMet(), "unmet db expectations")
}

func TestQuoteEngineImpl_GetQuoteSql(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	rows := pgxmock.NewRows([]string{"id", "data", "type"}).AddRow(13, "The purpose of knowledge is action, not knowledge", "quote")
	mockDb.ExpectQuery(`
SELECT id, data, type
FROM quotes
WHERE id =`).WithArgs(13).WillReturnRows(rows)
	engine := &QuoteEngineImpl{
		rng:    rand.New(&rand.PCGSource{}),
		lock:   sync.Mutex{},
		dbPool: mockDb,
		size:   25,
	}
	engine.rng.Seed(2222)
	expected := model.Quote{
		Id:   13,
		Data: "The purpose of knowledge is action, not knowledge",
		Type: "quote",
	}
	got, err := engine.GetQuoteSql(context.Background())
	assert.Nil(t, err, "got error")
	assert.Equal(t, expected, got, "Returned value does not match expected")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "Unmet db expectation")

}

// Verify that if the row is missing in underlying connection, scanned returns error from pgx
func TestQuoteEngineImpl_GetQuoteSqlMissingRow(t *testing.T) {

	mockDb, err := pgxmock.NewPool()
	require.Nil(t, err, "failed to build mock db")
	rows := pgxmock.NewRows([]string{"id", "data", "type"})
	mockDb.ExpectQuery(`
SELECT id, data, type
FROM quotes
WHERE id =`).WithArgs(13).WillReturnRows(rows)
	engine := &QuoteEngineImpl{
		rng:    rand.New(&rand.PCGSource{}),
		lock:   sync.Mutex{},
		dbPool: mockDb,
		size:   25,
	}
	engine.rng.Seed(2222)
	_, err = engine.GetQuoteSql(context.Background())
	assert.ErrorIs(t, err, pgx.ErrNoRows, "did not get missing result error")
	assert.Nil(t, mockDb.ExpectationsWereMet(), "unmet db expectations")

}

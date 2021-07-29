package util

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type SqliteConnectionPool struct {
	capacity uint
	pool     chan *sql.DB
	lock     *sync.WaitGroup
}

type StatementPool struct {
	capacity uint
	pool     chan *sql.Stmt
}

func NewSqliteConnectionPool(connString string, maxConns uint) (*SqliteConnectionPool, error) {
	pool := make(chan *sql.DB, maxConns)
	for i := 0; i < int(maxConns); i++ {
		conn, err := sql.Open("sqlite3", connString)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to connect to %s", connString)
		}
		pool <- conn
	}
	lock := &sync.WaitGroup{}
	return &SqliteConnectionPool{
		maxConns,
		pool,
		lock,
	}, nil
}

func (p *SqliteConnectionPool) Checkout() *sql.DB {
	p.lock.Wait() // block if there's a global lock on the pool
	return <-p.pool
}

func (p *SqliteConnectionPool) Checkin(conn *sql.DB) {
	p.pool <- conn
}

func (p *SqliteConnectionPool) Cleanup() error {
	for i := 0; i < int(p.capacity); i++ {
		err := p.Checkout().Close()
		if err != nil {
			return errors.Wrap(err, "failed to close connection")
		}
	}
	p.capacity = 0
	close(p.pool)
	return nil
}

func NewStatementPool(connPool *SqliteConnectionPool, query string) (*StatementPool, error) {
	pool := make(chan *sql.Stmt, connPool.capacity)
	connPool.lock.Add(1) // globally lock the connection pool to ensure 1 statement per connection
	defer connPool.lock.Done()
	for i := 0; i < int(connPool.capacity); i++ {
		conn := <-connPool.pool
		defer func() { connPool.pool <- conn }()
		stmt, err := conn.Prepare(query)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to prepare query %s", query)
		}
		pool <- stmt
	}
	return &StatementPool{
		connPool.capacity,
		pool,
	}, nil
}

func (p *StatementPool) Checkout() *sql.Stmt {
	return <-p.pool
}

func (p *StatementPool) Checkin(stmt *sql.Stmt) {
	p.pool <- stmt
}

func (p *StatementPool) Cleanup() error {
	for i := 0; i < int(p.capacity); i++ {
		err := p.Checkout().Close()
		if err != nil {
			return errors.Wrap(err, "failed to close statement")
		}
	}
	p.capacity = 0
	close(p.pool)
	return nil
}

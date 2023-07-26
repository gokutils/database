package database

import (
	"context"
	"database/sql"
)

type Scan interface {
	Scan(v ...interface{}) error
}

type Scanner func(row Scan) error

type QueryExecutor interface {
	ExecContext(ctx context.Context, scanner Scanner, sql string) error
	QueryRowContext(ctx context.Context, scanner Scanner, sql string, args ...interface{}) error
	QueryContext(ctx context.Context, scanner Scanner, sql string, args ...interface{}) error
}

type Queryer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type DatabaseInterface interface {
	Queryer
	BeginTx(ctx context.Context, sopts *sql.TxOptions) (*sql.Tx, error)
	Begin() (*sql.Tx, error)
}

type DB interface {
	QueryExecutor
	GetQueryerAndUnLocker(ctx context.Context) (Queryer, func())
	GetQueryerLocked(ctx context.Context, fn func(Queryer) error) error
	GetDatabaseLocked(ctx context.Context, fn func(db DatabaseInterface) error)
}

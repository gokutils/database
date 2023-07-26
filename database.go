package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/gokutils/txctx"
)

type NoOpBeginDb struct {
	*sql.Tx
}

func (l NoOpBeginDb) BeginTx(ctx context.Context, sopts *sql.TxOptions) (*sql.Tx, error) {
	return l.Tx, nil
}

func (l NoOpBeginDb) Begin() (*sql.Tx, error) {
	return l.Tx, nil
}

type LockerDB struct {
	db     *sql.DB
	tx     *sql.Tx
	lock   *sync.Mutex
	driver string
}

func (l *LockerDB) GetDatabaseAndUnLocker(ctx context.Context) (DatabaseInterface, func()) {
	if txctx.IsTxContext(ctx) {
		if v, ok := txctx.GetValue(ctx, l.db); ok {
			if tx, ok := v.(*LockerDB); ok {
				tx.lock.Lock()
				return NoOpBeginDb{Tx: tx.tx}, func() { tx.lock.Unlock() }
			} else {
				l.lock.Lock()
				return l.db, func() { l.lock.Unlock() }
			}
		} else {
			tx, _ := l.db.BeginTx(ctx, nil)
			tmp := &LockerDB{db: l.db, tx: tx, lock: &sync.Mutex{}}
			txctx.Add(ctx, tmp)
			txctx.SetValue(ctx, l.db, tmp)
			tmp.lock.Lock()
			return NoOpBeginDb{Tx: tmp.tx}, func() { tmp.lock.Unlock() }
		}
	} else {
		return l.db, func() {}
	}
}

func (l *LockerDB) GetQueryerAndUnLocker(ctx context.Context) (Queryer, func()) {
	return l.GetDatabaseAndUnLocker(ctx)
}

func (l *LockerDB) GetQueryerLocked(ctx context.Context, fn func(Queryer) error) error {
	db, lockFn := l.GetDatabaseAndUnLocker(ctx)
	defer lockFn()
	return fn(db)
}

/*
Ces fonction son la pour permetre d'utiliser les tx en multi thread
la db n'a pas se soucie la car ces un pool
*/
func (l *LockerDB) GetDatabaseLocked(ctx context.Context, fn func(db DatabaseInterface) error) error {
	db, lockFn := l.GetDatabaseAndUnLocker(ctx)
	defer lockFn()
	return fn(db)
}

func (l *LockerDB) BeginTx(ctx context.Context, sopts *sql.TxOptions) (*sql.Tx, error) {
	db, lockFn := l.GetDatabaseAndUnLocker(ctx)
	lockFn()
	if tx, ok := db.(NoOpBeginDb); ok {
		return tx.Tx, nil
	}
	return nil, fmt.Errorf("SQL BEGIN no tx ctx")
}

func (l *LockerDB) Begin() (*sql.Tx, error) {
	return nil, fmt.Errorf("SQL BEGIN no sql ctx")
}

func (l *LockerDB) Commit(ctx context.Context) error {
	return l.tx.Commit()
}

func (l *LockerDB) Rollback(ctx context.Context) error {
	return l.tx.Rollback()
}

func (l *LockerDB) GetDB() *sql.DB {
	return l.db
}

func (l *LockerDB) Driver() string {
	return l.driver
}

func (l *LockerDB) Close() {
	l.db.Close()
}

func NewLockerDB(db *sql.DB, driver string) DB {
	return &LockerDB{db: db, lock: &sync.Mutex{}, driver: driver}
}

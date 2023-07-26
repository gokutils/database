package database

import (
	"context"
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

func (r *LockerDB) ExecContext(ctx context.Context, sql string, args ...interface{}) error {
	return r.GetDatabaseLocked(ctx, func(db DatabaseInterface) error {
		_, err := db.ExecContext(ctx, sql, args...)
		return err
	})
}

func (r *LockerDB) QueryRowContext(ctx context.Context, scanner Scanner, sql string, args ...interface{}) error {
	return r.GetDatabaseLocked(ctx, func(db DatabaseInterface) error {
		row := db.QueryRowContext(ctx, sql, args...)
		if err := scanner(row); err != nil {
			return err
		}
		return nil
	})
}

func (r *LockerDB) QueryContext(ctx context.Context, scanner Scanner, sql string, args ...interface{}) error {
	return r.GetDatabaseLocked(ctx, func(db DatabaseInterface) error {
		rows, err := db.QueryContext(ctx, sql, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err = scanner(rows); err != nil {
				return err
			}
		}
		return nil
	})
}

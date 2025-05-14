package tx_manager

import (
	"context"
	"database/sql"
	"fmt"
)

type Beginner interface {
	Begin() (*sql.Tx, error)
}

type txKey struct{}

func WithTx(ctx context.Context, db any) (context.Context, error) {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return ctx, nil
	}

	b, ok := db.(Beginner)
	if ok {
		tx, err := b.Begin()
		if err != nil {
			return nil, fmt.Errorf("begin error: %w", err)
		}
		return context.WithValue(ctx, txKey{}, tx), nil
	}
	return ctx, nil
}

func Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit error: %w", err)
		}
	}
	return nil
}

func Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("rollback error: %w", err)
		}
	}
	return nil
}

func CommitOnDefer(ctx context.Context, err *error) {
	if *err != nil {
		if e := Rollback(ctx); e != nil {
			*err = fmt.Errorf("rollback error: %w; %w", e, *err)
		}
		return
	}
	if e := Commit(ctx); e != nil {
		*err = fmt.Errorf("commit error: %w", e)
	}
	return
}

type DBWithTx struct {
	*sql.DB
}

func NewDBWithTx(db *sql.DB) *DBWithTx {
	return &DBWithTx{db}
}

func (db *DBWithTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.QueryContext(ctx, query, args...)
	}
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DBWithTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DBWithTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	if ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return db.DB.ExecContext(ctx, query, args...)
}

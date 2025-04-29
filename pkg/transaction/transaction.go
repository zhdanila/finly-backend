package transaction

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type TransactionExecutor interface {
	WithTransaction(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error
}

type transactionExecutor struct{}

func (t *transactionExecutor) WithTransaction(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v; original error: %w", rbErr, err)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func NewTransactionExecutor() TransactionExecutor {
	return &transactionExecutor{}
}

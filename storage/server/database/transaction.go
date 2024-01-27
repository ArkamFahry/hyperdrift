package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction struct {
	db *pgxpool.Pool
}

func NewTransaction(db *pgxpool.Pool) *Transaction {
	return &Transaction{
		db: db,
	}
}

func (t *Transaction) WithTransaction(ctx context.Context, txnOptions *pgx.TxOptions, fn func(tx pgx.Tx) error) error {
	if txnOptions == nil {
		txnOptions = &pgx.TxOptions{}
	}

	tx, err := t.db.BeginTx(ctx, *txnOptions)
	if err != nil {
		return fmt.Errorf("begin error: %w", err)
	}

	err = fn(tx)

	if err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("rollback error: %v, transaction error: %w", rbErr, err)
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}

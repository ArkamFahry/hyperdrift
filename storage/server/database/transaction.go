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

func (t *Transaction) WithTransaction(ctx context.Context, fn func(*Queries) error) error {
	tx, err := t.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)

	err = fn(q)

	if err != nil {
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("rollback error: %v, transaction error: %w", rbErr, err)
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

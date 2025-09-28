package db

import (
	"context"
	"debt-manager/internal/contextkeys"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxRunner struct {
	pool *pgxpool.Pool
}

func NewTxRunner(pool *pgxpool.Pool) *TxRunner {
	return &TxRunner{pool: pool}
}

func (r *TxRunner) WithCtxUserTx(
	ctx context.Context,
	fn func(q *Queries) error,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT app.set_user($1)", ctx.Value(contextkeys.UserID{})); err != nil {
		return err
	}

	if err := fn(New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *TxRunner) WithTx(
	ctx context.Context,
	fn func(q *Queries) error,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxRunner struct {
	pool *pgxpool.Pool
}

func NewTxRunner(pool *pgxpool.Pool) *TxRunner {
	return &TxRunner{pool: pool}
}

func (r *TxRunner) WithUserTx(
	ctx context.Context,
	userID uuid.UUID,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SELECT app.set_user($1)", userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

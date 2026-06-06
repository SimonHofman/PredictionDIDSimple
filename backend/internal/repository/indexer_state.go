package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IndexerStateRepo struct {
	pool *pgxpool.Pool
}

func NewIndexerStateRepo(pool *pgxpool.Pool) *IndexerStateRepo {
	return &IndexerStateRepo{pool: pool}
}

func (r *IndexerStateRepo) GetLastBlock(ctx context.Context) (uint64, error) {
	var block uint64
	err := r.pool.QueryRow(ctx, `SELECT last_block FROM indexer_state WHERE id = 1`).Scan(&block)
	return block, err
}

func (r *IndexerStateRepo) SetLastBlock(ctx context.Context, block uint64, factory string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE indexer_state SET last_block = $2, factory_address = COALESCE($3, factory_address), updated_at = NOW()
		WHERE id = 1
	`, 1, block, factory)
	return err
}

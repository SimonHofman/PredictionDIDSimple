// Package repository 索引器状态仓储（记录已扫描区块）
package repository

// 导入依赖
import (
	"context" // 上下文

	"github.com/jackc/pgx/v5/pgxpool" // 连接池
)

// IndexerStateRepo 索引器状态仓储
type IndexerStateRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewIndexerStateRepo 创建索引器状态仓储
func NewIndexerStateRepo(pool *pgxpool.Pool) *IndexerStateRepo {
	return &IndexerStateRepo{pool: pool}
}

// GetLastBlock 获取上次处理的区块号
func (r *IndexerStateRepo) GetLastBlock(ctx context.Context) (uint64, error) {
	var block uint64
	err := r.pool.QueryRow(ctx, `SELECT last_block FROM indexer_state WHERE id = 1`).Scan(&block)
	return block, err
}

// SetLastBlock 更新已处理区块号
func (r *IndexerStateRepo) SetLastBlock(ctx context.Context, block uint64, factory string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE indexer_state SET last_block = $2, factory_address = COALESCE($3, factory_address), updated_at = NOW()
		WHERE id = 1
	`, 1, block, factory)
	return err
}

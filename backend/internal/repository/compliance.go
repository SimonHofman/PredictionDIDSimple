// Package repository 合规仓储（地理日志、KYC 事件）
package repository

// 导入依赖
import (
	"context" // 上下文

	"github.com/jackc/pgx/v5/pgxpool" // 连接池
)

// ComplianceRepo 合规仓储
type ComplianceRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewComplianceRepo 创建合规仓储
func NewComplianceRepo(pool *pgxpool.Pool) *ComplianceRepo {
	return &ComplianceRepo{pool: pool}
}

// LogGeo 记录地理访问日志
func (r *ComplianceRepo) LogGeo(ctx context.Context, ip, country, path string, allowed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO geo_access_log (ip, country_code, allowed, path) VALUES ($1,$2,$3,$4)
	`, ip, country, allowed, path)
	return err
}

// LogKYC 记录 KYC 回调事件
func (r *ComplianceRepo) LogKYC(ctx context.Context, externalID, userAddr, status string, raw []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_events (external_id, user_address, status, raw_json) VALUES ($1,$2,$3,$4::jsonb)
	`, externalID, userAddr, status, string(raw))
	return err
}

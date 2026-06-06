// Package repository 提供数据库持久层的仓储实现
package repository

// 导入依赖
import (
	"context" // 上下文，用于控制请求生命周期

	"github.com/jackc/pgx/v5/pgxpool" // PostgreSQL 连接池
)

// PlatformStats 平台统计数据结构体
// 包含交易数量、交易额、手续费、活跃用户等平台级指标
type PlatformStats struct {
	TradeCount    int64  `json:"trade_count"`    // 总交易次数
	TradeVolume   string `json:"trade_volume"`   // 总交易额
	FeesCollected string `json:"fees_collected"` // 已收取的手续费总额
	ActiveUsers   int64  `json:"active_users"`   // 活跃用户数
	OpenMarkets   int64  `json:"open_markets"`   // 当前开放的市场数量
	TVLApprox     string `json:"tvl_approx"`     // 总锁仓量（近似值）
}

// StatsRepo 统计数据仓储结构体，封装平台统计相关的数据库操作
type StatsRepo struct {
	pool *pgxpool.Pool // 数据库连接池实例
}

// NewStatsRepo 创建新的统计数据仓储实例
// 参数 pool: 数据库连接池
// 返回: StatsRepo 指针
func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{pool: pool} // 初始化并返回仓储实例
}

// Platform 查询平台级统计数据
// 通过多个子查询聚合平台的关键指标
// 返回: 平台统计数据和错误信息
func (r *StatsRepo) Platform(ctx context.Context) (*PlatformStats, error) {
	var s PlatformStats // 初始化统计数据对象
	// 执行聚合查询：统计交易数、交易额、手续费（0.3%费率）、活跃用户数、开放市场数、总锁仓量
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM trades),
			COALESCE((SELECT SUM(amount)::text FROM trades), '0'),
			COALESCE((SELECT SUM(amount) * 0.003 FROM trades)::text, '0'),
			(SELECT COUNT(DISTINCT user_address) FROM positions),
			(SELECT COUNT(*) FROM markets WHERE status = 'OPEN'),
			COALESCE((SELECT SUM(yes_pool + no_pool)::text FROM markets), '0')
	`).Scan(&s.TradeCount, &s.TradeVolume, &s.FeesCollected, &s.ActiveUsers, &s.OpenMarkets, &s.TVLApprox)
	if err != nil {
		return nil, err // 查询出错时返回错误
	}
	return &s, nil // 返回统计结果
}

// UpdateMarketPool 更新市场的资金池储备和价格
// 参数 marketID: 市场ID
// 参数 yes: Yes 池的储备金额
// 参数 no: No 池的储备金额
// 参数 priceBps: Yes 选项价格（基点表示，1 bps = 0.01%）
// 返回: 错误信息
func (r *StatsRepo) UpdateMarketPool(ctx context.Context, marketID int64, yes, no string, priceBps int) error {
	// 更新市场的 reserve_yes、reserve_no、price_yes_bps 字段
	_, err := r.pool.Exec(ctx, `
		UPDATE markets SET reserve_yes = $2::numeric, reserve_no = $3::numeric, price_yes_bps = $4, updated_at = NOW()
		WHERE id = $1
	`, marketID, yes, no, priceBps)
	return err // 返回执行结果的错误信息
}

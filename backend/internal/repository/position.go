// Package repository 提供数据库持久层的仓储实现
package repository

// 导入依赖
import (
	"context" // 上下文，用于控制请求生命周期

	"github.com/jackc/pgx/v5/pgxpool"                  // PostgreSQL 连接池
	"github.com/prediction-did/simple/internal/models" // 数据模型定义
)

// PositionRepo 仓位仓储结构体，封装用户持仓相关的数据库操作
type PositionRepo struct {
	pool *pgxpool.Pool // 数据库连接池实例
}

// NewPositionRepo 创建新的仓位仓储实例
// 参数 pool: 数据库连接池
// 返回: PositionRepo 指针
func NewPositionRepo(pool *pgxpool.Pool) *PositionRepo {
	return &PositionRepo{pool: pool} // 初始化并返回仓储实例
}

// AddTrade 添加交易记录并更新用户持仓
// 参数 marketID: 市场ID
// 参数 userAddress: 用户钱包地址
// 参数 outcome: 结果选项（0=Yes, 1=No）
// 参数 amount: 交易金额
// 返回: 错误信息
func (r *PositionRepo) AddTrade(ctx context.Context, marketID int64, userAddress string, outcome int, amount string) error {
	yesAmt := "0" // 初始化 Yes 金额为 0
	noAmt := "0"  // 初始化 No 金额为 0
	if outcome == 0 {
		// 如果选择 Yes 结果，设置 Yes 金额
		yesAmt = amount
	} else {
		// 如果选择 No 结果，设置 No 金额
		noAmt = amount
	}
	// 执行 SQL 插入或更新操作
	// 使用 ON CONFLICT 实现 upsert（存在则更新，不存在则插入）
	_, err := r.pool.Exec(ctx, `
		INSERT INTO positions (market_id, user_address, yes_amount, no_amount, updated_at)
		VALUES ($1, LOWER($2), $3::numeric, $4::numeric, NOW())
		ON CONFLICT (market_id, user_address) DO UPDATE SET
			yes_amount = positions.yes_amount + EXCLUDED.yes_amount,
			no_amount = positions.no_amount + EXCLUDED.no_amount,
			updated_at = NOW()
	`, marketID, userAddress, yesAmt, noAmt)
	return err // 返回执行结果的错误信息
}

// SetClaimed 设置用户持仓为已领取状态
// 参数 marketID: 市场ID
// 参数 userAddress: 用户钱包地址
// 返回: 错误信息
func (r *PositionRepo) SetClaimed(ctx context.Context, marketID int64, userAddress string) error {
	// 更新 claimed 字段为 true，匹配市场ID和用户地址
	_, err := r.pool.Exec(ctx, `
		UPDATE positions SET claimed = true, updated_at = NOW()
		WHERE market_id = $1 AND LOWER(user_address) = LOWER($2)
	`, marketID, userAddress)
	return err // 返回执行结果的错误信息
}

// ListByUser 查询指定用户的所有持仓记录，包含关联的市场信息
// 参数 address: 用户钱包地址
// 返回: 持仓列表和错误信息
func (r *PositionRepo) ListByUser(ctx context.Context, address string) ([]models.Position, error) {
	// 执行联合查询，关联 positions 和 markets 表
	// 按更新时间倒序排列
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.market_id, p.user_address, p.yes_amount::text, p.no_amount::text, p.claimed, p.updated_at,
		       m.id, m.match_id, m.chain_id, m.factory_address, m.market_address,
		       m.on_chain_market_id, m.match_ref, m.question, m.end_time, m.status,
		       m.winning_outcome, m.yes_pool::text, m.no_pool::text,
		       NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL
		FROM positions p
		JOIN markets m ON m.id = p.market_id
		WHERE LOWER(p.user_address) = LOWER($1)
		ORDER BY p.updated_at DESC
	`, address)
	if err != nil {
		return nil, err // 查询出错时返回错误
	}
	defer rows.Close() // 确保关闭结果集

	var out []models.Position // 初始化结果切片
	for rows.Next() {         // 遍历查询结果的每一行
		var p models.Position                       // 持仓对象
		var m models.Market                         // 市场对象
		var matchID, mtID *int64                    // 比赛ID相关的可空指针
		var mtExt, mtHome, mtAway, mtStatus *string // 比赛详情的可空字符串
		var mtKick interface{}                      // 开球时间（可空）
		var mtHomeScore, mtAwayScore *int           // 比分的可空指针
		// 扫描查询结果到变量中
		err := rows.Scan(
			&p.ID, &p.MarketID, &p.UserAddress, &p.YesAmount, &p.NoAmount, &p.Claimed, &p.UpdatedAt,
			&m.ID, &matchID, &m.ChainID, &m.FactoryAddress, &m.MarketAddress,
			&m.OnChainMarketID, &m.MatchRef, &m.Question, &m.EndTime, &m.Status,
			&m.WinningOutcome, &m.YesPool, &m.NoPool,
			&mtID, &mtExt, &mtHome, &mtAway, &mtKick, &mtStatus, &mtHomeScore, &mtAwayScore,
		)
		if err != nil {
			return nil, err // 扫描出错时返回错误
		}
		m.MatchID = matchID  // 设置市场关联的比赛ID
		p.Market = &m        // 将市场信息关联到持仓对象
		out = append(out, p) // 追加到结果切片
	}
	return out, rows.Err() // 返回结果并检查遍历是否有错误
}

// InsertTrade 插入交易记录到 trades 表
// 使用 ON CONFLICT DO NOTHING 避免重复插入（幂等操作）
// 参数 marketID: 市场ID
// 参数 txHash: 交易哈希
// 参数 logIndex: 日志索引
// 参数 blockNumber: 区块号
// 参数 userAddress: 用户钱包地址
// 参数 outcome: 结果选项
// 参数 amount: 交易金额
// 返回: 错误信息
func (r *PositionRepo) InsertTrade(ctx context.Context, marketID int64, txHash string, logIndex int, blockNumber int64, userAddress string, outcome int, amount string) error {
	// 插入交易记录，如果交易哈希和日志索引已存在则忽略
	_, err := r.pool.Exec(ctx, `
		INSERT INTO trades (market_id, tx_hash, log_index, block_number, user_address, outcome, amount)
		VALUES ($1, $2, $3, $4, LOWER($5), $6, $7::numeric)
		ON CONFLICT (tx_hash, log_index) DO NOTHING
	`, marketID, txHash, logIndex, blockNumber, userAddress, outcome, amount)
	return err // 返回执行结果的错误信息
}

// Package repository Oracle 任务仓储
package repository

// 导入依赖
import (
	"context" // 上下文
	"time"    // 时间

	"github.com/jackc/pgx/v5/pgxpool" // 连接池
)

// OracleJob Oracle 任务数据模型
type OracleJob struct {
	ID              int64     `json:"id"`                         // 主键
	MatchID         *int64    `json:"match_id,omitempty"`         // 关联比赛 ID
	MarketID        int64     `json:"market_id"`                  // 关联市场 ID
	Status          string    `json:"status"`                     // 状态
	PrimaryHome     *int      `json:"primary_home,omitempty"`     // 主源主队得分
	PrimaryAway     *int      `json:"primary_away,omitempty"`     // 主源客队得分
	SecondaryHome   *int      `json:"secondary_home,omitempty"`   // 备源主队得分
	SecondaryAway   *int      `json:"secondary_away,omitempty"`   // 备源客队得分
	ProposedOutcome *int      `json:"proposed_outcome,omitempty"` // 提议结果
	TxHash          *string   `json:"tx_hash,omitempty"`          // 交易哈希
	ErrorMessage    *string   `json:"error_message,omitempty"`    // 错误消息
	ExecuteAfter    time.Time `json:"execute_after"`              // 最早执行时间
	CreatedAt       time.Time `json:"created_at"`                 // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`                 // 更新时间
	MarketAddress   string    `json:"market_address,omitempty"`   // 市场合约地址（JOIN 取出）
	Question        string    `json:"question,omitempty"`         // 问题（JOIN 取出）
}

// OracleJobRepo Oracle 任务仓储
type OracleJobRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewOracleJobRepo 创建 Oracle 任务仓储
func NewOracleJobRepo(pool *pgxpool.Pool) *OracleJobRepo {
	return &OracleJobRepo{pool: pool}
}

// Create 创建新任务
func (r *OracleJobRepo) Create(ctx context.Context, marketID int64, matchID *int64, executeAfter time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO oracle_jobs (market_id, match_id, status, execute_after)
		VALUES ($1, $2, 'pending', $3)
		RETURNING id
	`, marketID, matchID, executeAfter).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// HasActiveForMarket 判断某市场是否有活跃任务
func (r *OracleJobRepo) HasActiveForMarket(ctx context.Context, marketID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM oracle_jobs
			WHERE market_id = $1 AND status IN ('pending','submitted','manual_review')
		)
	`, marketID).Scan(&exists)
	return exists, err
}

// ListDue 获取所有到期待处理任务
func (r *OracleJobRepo) ListDue(ctx context.Context, now time.Time) ([]OracleJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT j.id, j.match_id, j.market_id, j.status,
			j.primary_home, j.primary_away, j.secondary_home, j.secondary_away,
			j.proposed_outcome, j.tx_hash, j.error_message, j.execute_after, j.created_at, j.updated_at,
			m.market_address, m.question
		FROM oracle_jobs j
		JOIN markets m ON m.id = j.market_id
		WHERE j.status = 'pending' AND j.execute_after <= $1
		ORDER BY j.execute_after ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OracleJob
	for rows.Next() {
		j, err := scanOracleJobRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// ListAll 列出所有任务（可选 status 过滤）
func (r *OracleJobRepo) ListAll(ctx context.Context, status string, limit int) ([]OracleJob, error) {
	if limit <= 0 {
		limit = 50 // 默认 50
	}
	q := `
		SELECT j.id, j.match_id, j.market_id, j.status,
			j.primary_home, j.primary_away, j.secondary_home, j.secondary_away,
			j.proposed_outcome, j.tx_hash, j.error_message, j.execute_after, j.created_at, j.updated_at,
			m.market_address, m.question
		FROM oracle_jobs j
		JOIN markets m ON m.id = j.market_id`
	args := []interface{}{limit}
	// 可选状态过滤
	if status != "" {
		q += ` WHERE j.status = $2`
		args = []interface{}{limit, status}
	}
	q += ` ORDER BY j.updated_at DESC LIMIT $1`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OracleJob
	for rows.Next() {
		j, err := scanOracleJobRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// UpdateStatus 更新任务状态及附带字段
func (r *OracleJobRepo) UpdateStatus(ctx context.Context, id int64, status string, fields map[string]interface{}) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE oracle_jobs SET status = $2,
			primary_home = COALESCE($3, primary_home),
			primary_away = COALESCE($4, primary_away),
			secondary_home = COALESCE($5, secondary_home),
			secondary_away = COALESCE($6, secondary_away),
			proposed_outcome = COALESCE($7, proposed_outcome),
			tx_hash = COALESCE($8, tx_hash),
			error_message = COALESCE($9, error_message),
			updated_at = NOW()
		WHERE id = $1
	`, id, status,
		fields["primary_home"], fields["primary_away"],
		fields["secondary_home"], fields["secondary_away"],
		fields["proposed_outcome"], fields["tx_hash"], fields["error_message"])
	return err
}

// scanOracleJobRow 从行扫描 OracleJob
func scanOracleJobRow(rows interface {
	Scan(dest ...any) error
}) (OracleJob, error) {
	var j OracleJob
	var txHash, errMsg *string
	err := rows.Scan(
		&j.ID, &j.MatchID, &j.MarketID, &j.Status,
		&j.PrimaryHome, &j.PrimaryAway, &j.SecondaryHome, &j.SecondaryAway,
		&j.ProposedOutcome, &txHash, &errMsg, &j.ExecuteAfter, &j.CreatedAt, &j.UpdatedAt,
		&j.MarketAddress, &j.Question,
	)
	j.TxHash = txHash       // 交易哈希
	j.ErrorMessage = errMsg // 错误信息
	return j, err
}

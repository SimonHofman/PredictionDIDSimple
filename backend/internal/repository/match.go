// Package repository 比赛仓储
package repository

// 导入依赖
import (
	"context" // 上下文
	"fmt"     // 格式化
	"strings" // 字符串
	"time"    // 时间

	"github.com/jackc/pgx/v5/pgxpool"                  // 连接池
	"github.com/prediction-did/simple/internal/models" // 数据模型
)

// MatchRepo 比赛仓储
type MatchRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewMatchRepo 创建比赛仓储
func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{pool: pool}
}

// List 分页查询比赛列表（可选 status 过滤）
func (r *MatchRepo) List(ctx context.Context, status string, limit, offset int) ([]models.Match, error) {
	if limit <= 0 {
		limit = 20 // 默认 20
	}
	// 基础 SQL
	q := `SELECT id, external_id, home_team, away_team, kickoff_at, status, home_score, away_score
	      FROM matches`
	args := []interface{}{}
	where := []string{}
	// 动态 WHERE
	if status != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, status)
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	// 排序与分页
	q += fmt.Sprintf(" ORDER BY kickoff_at ASC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Match
	for rows.Next() {
		var m models.Match
		// 扫描行
		if err := rows.Scan(&m.ID, &m.ExternalID, &m.HomeTeam, &m.AwayTeam, &m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// GetByID 根据 ID 获取比赛
func (r *MatchRepo) GetByID(ctx context.Context, id int64) (*models.Match, error) {
	var m models.Match
	err := r.pool.QueryRow(ctx,
		`SELECT id, external_id, home_team, away_team, kickoff_at, status, home_score, away_score
		 FROM matches WHERE id = $1`, id,
	).Scan(&m.ID, &m.ExternalID, &m.HomeTeam, &m.AwayTeam, &m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Upsert 插入或更新比赛（按 external_id 去重）
func (r *MatchRepo) Upsert(ctx context.Context, m models.Match) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO matches (external_id, home_team, away_team, kickoff_at, status, home_score, away_score, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (external_id) DO UPDATE SET
			home_team = EXCLUDED.home_team,
			away_team = EXCLUDED.away_team,
			kickoff_at = EXCLUDED.kickoff_at,
			status = EXCLUDED.status,
			home_score = EXCLUDED.home_score,
			away_score = EXCLUDED.away_score,
			updated_at = NOW()
	`, m.ExternalID, m.HomeTeam, m.AwayTeam, m.KickoffAt, m.Status, m.HomeScore, m.AwayScore)
	return err
}

// SetStatus 更新比赛状态
func (r *MatchRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE matches SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

// GetByExternalID 根据外部 ID 查询比赛
func (r *MatchRepo) GetByExternalID(ctx context.Context, externalID string) (*models.Match, error) {
	var m models.Match
	err := r.pool.QueryRow(ctx,
		`SELECT id, external_id, home_team, away_team, kickoff_at, status, home_score, away_score
		 FROM matches WHERE external_id = $1`, externalID,
	).Scan(&m.ID, &m.ExternalID, &m.HomeTeam, &m.AwayTeam, &m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ParseKickoff 解析 RFC3339 格式的开球时间字符串
func ParseKickoff(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

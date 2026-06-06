// Package repository 市场仓储
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

// MarketRepo 市场仓储
type MarketRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewMarketRepo 创建市场仓储
func NewMarketRepo(pool *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{pool: pool}
}

// List 分页查询市场列表（可选 status 过滤）
func (r *MarketRepo) List(ctx context.Context, status string, limit, offset int) ([]models.Market, error) {
	if limit <= 0 {
		limit = 20 // 默认 20
	}
	// 构造 SQL
	q := `
		SELECT m.id, m.match_id, m.chain_id, m.factory_address, m.market_address,
		       m.on_chain_market_id, m.match_ref, m.question, m.end_time, m.status,
		       m.winning_outcome, m.yes_pool::text, m.no_pool::text,
		       COALESCE(m.market_type, 'BINARY'), COALESCE(m.outcome_count, 2), COALESCE(m.fee_bps, 30),
		       COALESCE(m.reserve_yes::text, '0'), COALESCE(m.reserve_no::text, '0'),
		       COALESCE(m.price_yes_bps::text, ''),
		       COALESCE(m.requires_vc, false), COALESCE(m.restricted_region, ''),
		       COALESCE(m.resolution_rule, 'HOME_WIN'),
		       mt.id, mt.external_id, mt.home_team, mt.away_team, mt.kickoff_at, mt.status,
		       mt.home_score, mt.away_score
		FROM markets m
		LEFT JOIN matches mt ON mt.id = m.match_id`
	args := []interface{}{}
	where := []string{}
	// 动态添加过滤条件
	if status != "" {
		where = append(where, fmt.Sprintf("m.status = $%d", len(args)+1))
		args = append(args, status)
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	// 排序与分页
	q += fmt.Sprintf(" ORDER BY m.end_time ASC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Market
	for rows.Next() {
		mk, err := scanMarketWithMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, mk)
	}
	return out, rows.Err()
}

// GetByID 根据 ID 获取单个市场
func (r *MarketRepo) GetByID(ctx context.Context, id int64) (*models.Market, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT m.id, m.match_id, m.chain_id, m.factory_address, m.market_address,
		       m.on_chain_market_id, m.match_ref, m.question, m.end_time, m.status,
		       m.winning_outcome, m.yes_pool::text, m.no_pool::text,
		       COALESCE(m.market_type, 'BINARY'), COALESCE(m.outcome_count, 2), COALESCE(m.fee_bps, 30),
		       COALESCE(m.reserve_yes::text, '0'), COALESCE(m.reserve_no::text, '0'),
		       COALESCE(m.price_yes_bps::text, ''),
		       COALESCE(m.requires_vc, false), COALESCE(m.restricted_region, ''),
		       COALESCE(m.resolution_rule, 'HOME_WIN'),
		       mt.id, mt.external_id, mt.home_team, mt.away_team, mt.kickoff_at, mt.status,
		       mt.home_score, mt.away_score
		FROM markets m
		LEFT JOIN matches mt ON mt.id = m.match_id
		WHERE m.id = $1`, id)
	return scanMarketRow(row)
}

// GetByAddress 根据合约地址获取市场
func (r *MarketRepo) GetByAddress(ctx context.Context, addr string) (*models.Market, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT m.id, m.match_id, m.chain_id, m.factory_address, m.market_address,
		       m.on_chain_market_id, m.match_ref, m.question, m.end_time, m.status,
		       m.winning_outcome, m.yes_pool::text, m.no_pool::text,
		       COALESCE(m.market_type, 'BINARY'), COALESCE(m.outcome_count, 2), COALESCE(m.fee_bps, 30),
		       COALESCE(m.reserve_yes::text, '0'), COALESCE(m.reserve_no::text, '0'),
		       COALESCE(m.price_yes_bps::text, ''),
		       COALESCE(m.requires_vc, false), COALESCE(m.restricted_region, ''),
		       COALESCE(m.resolution_rule, 'HOME_WIN'),
		       mt.id, mt.external_id, mt.home_team, mt.away_team, mt.kickoff_at, mt.status,
		       mt.home_score, mt.away_score
		FROM markets m
		LEFT JOIN matches mt ON mt.id = m.match_id
		WHERE LOWER(m.market_address) = LOWER($1)`, addr)
	return scanMarketRow(row)
}

// InsertFromChain 从链上事件插入市场（冲突时更新）
func (r *MarketRepo) InsertFromChain(ctx context.Context, mk models.Market) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO markets (
			match_id, chain_id, factory_address, market_address, on_chain_market_id,
			match_ref, question, end_time, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (market_address) DO UPDATE SET
			match_id = COALESCE(EXCLUDED.match_id, markets.match_id),
			updated_at = NOW()
		RETURNING id
	`, mk.MatchID, mk.ChainID, mk.FactoryAddress, mk.MarketAddress, mk.OnChainMarketID,
		mk.MatchRef, mk.Question, mk.EndTime, mk.Status).Scan(&id)
	return id, err
}

// UpdateResolved 更新市场为已结算
func (r *MarketRepo) UpdateResolved(ctx context.Context, marketAddress string, outcome int, yesPool, noPool string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE markets SET status = 'RESOLVED', winning_outcome = $2,
			yes_pool = $3::numeric, no_pool = $4::numeric, updated_at = NOW()
		WHERE LOWER(market_address) = LOWER($1)
	`, marketAddress, outcome, yesPool, noPool)
	return err
}

// ListOpenByMatchID 查询某比赛下所有 OPEN 状态的市场
func (r *MarketRepo) ListOpenByMatchID(ctx context.Context, matchID int64) ([]models.Market, error) {
	all, err := r.List(ctx, "OPEN", 200, 0)
	if err != nil {
		return nil, err
	}
	var out []models.Market
	for _, mk := range all {
		if mk.MatchID != nil && *mk.MatchID == matchID {
			full, err := r.GetByID(ctx, mk.ID)
			if err != nil {
				return nil, err
			}
			out = append(out, *full)
		}
	}
	return out, nil
}

// AdminMarketUpdate 管理员更新市场的参数
type AdminMarketUpdate struct {
	MatchID          int64  // 比赛 ID
	RequiresVC       bool   // 是否需要 VC
	RestrictedRegion string // 地区限制
	ResolutionRule   string // 结算规则
}

// RegisterAdmin 管理员更新市场属性
func (r *MarketRepo) RegisterAdmin(ctx context.Context, req AdminMarketUpdate) error {
	rule := req.ResolutionRule
	if rule == "" {
		rule = "HOME_WIN" // 默认规则
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE markets SET
			requires_vc = $2,
			restricted_region = NULLIF($3, ''),
			resolution_rule = $4,
			updated_at = NOW()
		WHERE match_id = $1
	`, req.MatchID, req.RequiresVC, req.RestrictedRegion, rule)
	return err
}

// SetVoid 设置市场为作废状态
func (r *MarketRepo) SetVoid(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE markets SET status = 'VOID', updated_at = NOW() WHERE id = $1`, id)
	return err
}

// UpdatePools 更新市场资金池
func (r *MarketRepo) UpdatePools(ctx context.Context, marketAddress string, yesPool, noPool string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE markets SET yes_pool = $2::numeric, no_pool = $3::numeric, updated_at = NOW()
		WHERE LOWER(market_address) = LOWER($1)
	`, marketAddress, yesPool, noPool)
	return err
}

// scanMarketWithMatch 从行扫描市场（带关联比赛）
func scanMarketWithMatch(rows interface {
	Next() bool
	Scan(dest ...any) error
}) (models.Market, error) {
	mk, err := scanMarketRow(rows)
	if err != nil {
		return models.Market{}, err
	}
	return *mk, nil
}

// scanMarketRow 从单行扫描市场
func scanMarketRow(row interface {
	Scan(dest ...any) error
}) (*models.Market, error) {
	var m models.Market
	var matchID, mtID *int64
	var mtExt, mtHome, mtAway, mtStatus *string
	var restricted *string
	var mtKick *time.Time
	var mtHomeScore, mtAwayScore *int
	// 扫描所有字段
	err := row.Scan(
		&m.ID, &matchID, &m.ChainID, &m.FactoryAddress, &m.MarketAddress,
		&m.OnChainMarketID, &m.MatchRef, &m.Question, &m.EndTime, &m.Status,
		&m.WinningOutcome, &m.YesPool, &m.NoPool,
		&m.MarketType, &m.OutcomeCount, &m.FeeBps,
		&m.ReserveYes, &m.ReserveNo, &m.PriceYesBps,
		&m.RequiresVC, &restricted, &m.ResolutionRule,
		&mtID, &mtExt, &mtHome, &mtAway, &mtKick, &mtStatus, &mtHomeScore, &mtAwayScore,
	)
	if err != nil {
		return nil, err
	}
	m.MatchID = matchID
	if restricted != nil {
		m.RestrictedRegion = *restricted
	}
	// 填充关联比赛
	if mtID != nil {
		m.Match = &models.Match{
			ID:         *mtID,
			ExternalID: derefStr(mtExt),
			HomeTeam:   derefStr(mtHome),
			AwayTeam:   derefStr(mtAway),
			KickoffAt:  derefTime(mtKick),
			Status:     derefStr(mtStatus),
			HomeScore:  mtHomeScore,
			AwayScore:  mtAwayScore,
		}
	}
	return &m, nil
}

// derefStr 解引用字符串指针
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefTime 解引用时间指针
func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

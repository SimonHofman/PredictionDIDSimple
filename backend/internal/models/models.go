// Package models 定义核心数据模型（比赛、市场、持仓、用户）
package models

import "time" // 时间类型

// Match 比赛实体
type Match struct {
	ID         int64     `json:"id"`                   // 主键
	ExternalID string    `json:"external_id"`          // 外部唯一标识（如 wc-2026-semi-001）
	HomeTeam   string    `json:"home_team"`            // 主队名称
	AwayTeam   string    `json:"away_team"`            // 客队名称
	KickoffAt  time.Time `json:"kickoff_at"`           // 开球时间
	Status     string    `json:"status"`               // 状态（SCHEDULED/LIVE/FINISHED/RESOLVED 等）
	HomeScore  *int      `json:"home_score,omitempty"` // 主队得分
	AwayScore  *int      `json:"away_score,omitempty"` // 客队得分
}

// Market 预测市场实体
type Market struct {
	ID               int64     `json:"id"`                          // 主键
	MatchID          *int64    `json:"match_id,omitempty"`          // 关联比赛 ID
	ChainID          int64     `json:"chain_id"`                    // 链 ID
	FactoryAddress   string    `json:"factory_address"`             // 工厂合约地址
	MarketAddress    string    `json:"market_address"`              // 市场合约地址
	OnChainMarketID  int64     `json:"on_chain_market_id"`          // 链上市场 ID
	MatchRef         string    `json:"match_ref"`                   // 比赛引用哈希
	Question         string    `json:"question"`                    // 预测问题
	EndTime          time.Time `json:"end_time"`                    // 结束时间
	Status           string    `json:"status"`                      // 市场状态
	WinningOutcome   *int      `json:"winning_outcome,omitempty"`   // 获胜结果
	YesPool          string    `json:"yes_pool"`                    // YES 资金池
	NoPool           string    `json:"no_pool"`                     // NO 资金池
	MarketType       string    `json:"market_type"`                 // 市场类型（binary/multi）
	OutcomeCount     int       `json:"outcome_count"`               // 结果数量
	FeeBps           int       `json:"fee_bps"`                     // 手续费基点
	ReserveYes       string    `json:"reserve_yes,omitempty"`       // YES 储备
	ReserveNo        string    `json:"reserve_no,omitempty"`        // NO 储备
	PriceYesBps      string    `json:"price_yes_bps,omitempty"`     // YES 价格基点
	RequiresVC       bool      `json:"requires_vc"`                 // 是否需要 VC 门控
	RestrictedRegion string    `json:"restricted_region,omitempty"` // 地区限制
	ResolutionRule   string    `json:"resolution_rule,omitempty"`   // 结算规则
	Match            *Match    `json:"match,omitempty"`             // 关联比赛（可选填充）
}

// Position 用户持仓实体
type Position struct {
	ID          int64     `json:"id"`               // 主键
	MarketID    int64     `json:"market_id"`        // 市场 ID
	UserAddress string    `json:"user_address"`     // 用户地址
	YesAmount   string    `json:"yes_amount"`       // YES 持有量
	NoAmount    string    `json:"no_amount"`        // NO 持有量
	Claimed     bool      `json:"claimed"`          // 是否已领取
	Market      *Market   `json:"market,omitempty"` // 关联市场（可选填充）
	UpdatedAt   time.Time `json:"updated_at"`       // 更新时间
}

// User 用户实体
type User struct {
	ID      int64   `json:"id"`            // 主键
	Address string  `json:"address"`       // 钱包地址
	DID     *string `json:"did,omitempty"` // 绑定的 DID
}

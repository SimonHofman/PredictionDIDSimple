package models

import "time"

type Match struct {
	ID         int64      `json:"id"`
	ExternalID string     `json:"external_id"`
	HomeTeam   string     `json:"home_team"`
	AwayTeam   string     `json:"away_team"`
	KickoffAt  time.Time  `json:"kickoff_at"`
	Status     string     `json:"status"`
	HomeScore  *int       `json:"home_score,omitempty"`
	AwayScore  *int       `json:"away_score,omitempty"`
}

type Market struct {
	ID              int64      `json:"id"`
	MatchID         *int64     `json:"match_id,omitempty"`
	ChainID         int64      `json:"chain_id"`
	FactoryAddress  string     `json:"factory_address"`
	MarketAddress   string     `json:"market_address"`
	OnChainMarketID int64      `json:"on_chain_market_id"`
	MatchRef        string     `json:"match_ref"`
	Question        string     `json:"question"`
	EndTime         time.Time  `json:"end_time"`
	Status          string     `json:"status"`
	WinningOutcome  *int       `json:"winning_outcome,omitempty"`
	YesPool          string     `json:"yes_pool"`
	NoPool           string     `json:"no_pool"`
	MarketType       string     `json:"market_type"`
	OutcomeCount     int        `json:"outcome_count"`
	FeeBps           int        `json:"fee_bps"`
	ReserveYes       string     `json:"reserve_yes,omitempty"`
	ReserveNo        string     `json:"reserve_no,omitempty"`
	PriceYesBps      string     `json:"price_yes_bps,omitempty"`
	RequiresVC       bool       `json:"requires_vc"`
	RestrictedRegion string     `json:"restricted_region,omitempty"`
	ResolutionRule   string     `json:"resolution_rule,omitempty"`
	Match            *Match     `json:"match,omitempty"`
}

type Position struct {
	ID            int64     `json:"id"`
	MarketID      int64     `json:"market_id"`
	UserAddress   string    `json:"user_address"`
	YesAmount     string    `json:"yes_amount"`
	NoAmount      string    `json:"no_amount"`
	Claimed       bool      `json:"claimed"`
	Market        *Market   `json:"market,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type User struct {
	ID      int64   `json:"id"`
	Address string  `json:"address"`
	DID     *string `json:"did,omitempty"`
}

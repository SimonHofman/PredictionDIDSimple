package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlatformStats struct {
	TradeCount    int64  `json:"trade_count"`
	TradeVolume   string `json:"trade_volume"`
	FeesCollected string `json:"fees_collected"`
	ActiveUsers   int64  `json:"active_users"`
	OpenMarkets   int64  `json:"open_markets"`
	TVLApprox     string `json:"tvl_approx"`
}

type StatsRepo struct {
	pool *pgxpool.Pool
}

func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{pool: pool}
}

func (r *StatsRepo) Platform(ctx context.Context) (*PlatformStats, error) {
	var s PlatformStats
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
		return nil, err
	}
	return &s, nil
}

func (r *StatsRepo) UpdateMarketPool(ctx context.Context, marketID int64, yes, no string, priceBps int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE markets SET reserve_yes = $2::numeric, reserve_no = $3::numeric, price_yes_bps = $4, updated_at = NOW()
		WHERE id = $1
	`, marketID, yes, no, priceBps)
	return err
}

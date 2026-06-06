package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/models"
)

type PositionRepo struct {
	pool *pgxpool.Pool
}

func NewPositionRepo(pool *pgxpool.Pool) *PositionRepo {
	return &PositionRepo{pool: pool}
}

func (r *PositionRepo) AddTrade(ctx context.Context, marketID int64, userAddress string, outcome int, amount string) error {
	yesAmt := "0"
	noAmt := "0"
	if outcome == 0 {
		yesAmt = amount
	} else {
		noAmt = amount
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO positions (market_id, user_address, yes_amount, no_amount, updated_at)
		VALUES ($1, LOWER($2), $3::numeric, $4::numeric, NOW())
		ON CONFLICT (market_id, user_address) DO UPDATE SET
			yes_amount = positions.yes_amount + EXCLUDED.yes_amount,
			no_amount = positions.no_amount + EXCLUDED.no_amount,
			updated_at = NOW()
	`, marketID, userAddress, yesAmt, noAmt)
	return err
}

func (r *PositionRepo) SetClaimed(ctx context.Context, marketID int64, userAddress string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE positions SET claimed = true, updated_at = NOW()
		WHERE market_id = $1 AND LOWER(user_address) = LOWER($2)
	`, marketID, userAddress)
	return err
}

func (r *PositionRepo) ListByUser(ctx context.Context, address string) ([]models.Position, error) {
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
		return nil, err
	}
	defer rows.Close()

	var out []models.Position
	for rows.Next() {
		var p models.Position
		var m models.Market
		var matchID, mtID *int64
		var mtExt, mtHome, mtAway, mtStatus *string
		var mtKick interface{}
		var mtHomeScore, mtAwayScore *int
		err := rows.Scan(
			&p.ID, &p.MarketID, &p.UserAddress, &p.YesAmount, &p.NoAmount, &p.Claimed, &p.UpdatedAt,
			&m.ID, &matchID, &m.ChainID, &m.FactoryAddress, &m.MarketAddress,
			&m.OnChainMarketID, &m.MatchRef, &m.Question, &m.EndTime, &m.Status,
			&m.WinningOutcome, &m.YesPool, &m.NoPool,
			&mtID, &mtExt, &mtHome, &mtAway, &mtKick, &mtStatus, &mtHomeScore, &mtAwayScore,
		)
		if err != nil {
			return nil, err
		}
		m.MatchID = matchID
		p.Market = &m
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PositionRepo) InsertTrade(ctx context.Context, marketID int64, txHash string, logIndex int, blockNumber int64, userAddress string, outcome int, amount string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO trades (market_id, tx_hash, log_index, block_number, user_address, outcome, amount)
		VALUES ($1, $2, $3, $4, LOWER($5), $6, $7::numeric)
		ON CONFLICT (tx_hash, log_index) DO NOTHING
	`, marketID, txHash, logIndex, blockNumber, userAddress, outcome, amount)
	return err
}

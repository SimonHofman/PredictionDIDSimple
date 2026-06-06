package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OracleJob struct {
	ID              int64      `json:"id"`
	MatchID         *int64     `json:"match_id,omitempty"`
	MarketID        int64      `json:"market_id"`
	Status          string     `json:"status"`
	PrimaryHome     *int       `json:"primary_home,omitempty"`
	PrimaryAway     *int       `json:"primary_away,omitempty"`
	SecondaryHome   *int       `json:"secondary_home,omitempty"`
	SecondaryAway   *int       `json:"secondary_away,omitempty"`
	ProposedOutcome *int       `json:"proposed_outcome,omitempty"`
	TxHash          *string    `json:"tx_hash,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	ExecuteAfter    time.Time  `json:"execute_after"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	MarketAddress   string     `json:"market_address,omitempty"`
	Question        string     `json:"question,omitempty"`
}

type OracleJobRepo struct {
	pool *pgxpool.Pool
}

func NewOracleJobRepo(pool *pgxpool.Pool) *OracleJobRepo {
	return &OracleJobRepo{pool: pool}
}

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

func (r *OracleJobRepo) ListAll(ctx context.Context, status string, limit int) ([]OracleJob, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `
		SELECT j.id, j.match_id, j.market_id, j.status,
			j.primary_home, j.primary_away, j.secondary_home, j.secondary_away,
			j.proposed_outcome, j.tx_hash, j.error_message, j.execute_after, j.created_at, j.updated_at,
			m.market_address, m.question
		FROM oracle_jobs j
		JOIN markets m ON m.id = j.market_id`
	args := []interface{}{limit}
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
	j.TxHash = txHash
	j.ErrorMessage = errMsg
	return j, err
}

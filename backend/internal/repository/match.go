package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/models"
)

type MatchRepo struct {
	pool *pgxpool.Pool
}

func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{pool: pool}
}

func (r *MatchRepo) List(ctx context.Context, status string, limit, offset int) ([]models.Match, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT id, external_id, home_team, away_team, kickoff_at, status, home_score, away_score
	      FROM matches`
	args := []interface{}{}
	where := []string{}
	if status != "" {
		where = append(where, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, status)
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
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
		if err := rows.Scan(&m.ID, &m.ExternalID, &m.HomeTeam, &m.AwayTeam, &m.KickoffAt, &m.Status, &m.HomeScore, &m.AwayScore); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

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

func (r *MatchRepo) SetStatus(ctx context.Context, id int64, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE matches SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	return err
}

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

func ParseKickoff(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

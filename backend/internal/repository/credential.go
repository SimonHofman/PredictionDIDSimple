package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Credential struct {
	ID             int64           `json:"id"`
	UserAddress    string          `json:"user_address"`
	CredentialType string          `json:"credential_type"`
	VCJSON         json.RawMessage `json:"vc_json"`
	ExpiresAt      time.Time       `json:"expires_at"`
	Revoked        bool            `json:"revoked"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CredentialRepo struct {
	pool *pgxpool.Pool
}

func NewCredentialRepo(pool *pgxpool.Pool) *CredentialRepo {
	return &CredentialRepo{pool: pool}
}

func (r *CredentialRepo) Insert(ctx context.Context, c Credential) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO credentials (user_address, credential_type, vc_json, expires_at)
		VALUES (LOWER($1), $2, $3, $4)
		RETURNING id
	`, c.UserAddress, c.CredentialType, c.VCJSON, c.ExpiresAt).Scan(&id)
	return id, err
}

func (r *CredentialRepo) ListByUser(ctx context.Context, address string) ([]Credential, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_address, credential_type, vc_json, expires_at, revoked, created_at
		FROM credentials
		WHERE LOWER(user_address) = LOWER($1) AND revoked = false AND expires_at > NOW()
		ORDER BY created_at DESC
	`, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Credential
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.ID, &c.UserAddress, &c.CredentialType, &c.VCJSON, &c.ExpiresAt, &c.Revoked, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CredentialRepo) HasValidType(ctx context.Context, address, credType string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM credentials
			WHERE LOWER(user_address) = LOWER($1)
			  AND credential_type = $2
			  AND revoked = false
			  AND expires_at > NOW()
		)
	`, address, credType).Scan(&ok)
	return ok, err
}

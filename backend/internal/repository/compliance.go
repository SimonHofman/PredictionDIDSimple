package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ComplianceRepo struct {
	pool *pgxpool.Pool
}

func NewComplianceRepo(pool *pgxpool.Pool) *ComplianceRepo {
	return &ComplianceRepo{pool: pool}
}

func (r *ComplianceRepo) LogGeo(ctx context.Context, ip, country, path string, allowed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO geo_access_log (ip, country_code, allowed, path) VALUES ($1,$2,$3,$4)
	`, ip, country, allowed, path)
	return err
}

func (r *ComplianceRepo) LogKYC(ctx context.Context, externalID, userAddr, status string, raw []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_events (external_id, user_address, status, raw_json) VALUES ($1,$2,$3,$4::jsonb)
	`, externalID, userAddr, status, string(raw))
	return err
}

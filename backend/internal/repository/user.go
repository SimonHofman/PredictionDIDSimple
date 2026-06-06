package repository

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/models"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) UpsertByAddress(ctx context.Context, address string) (*models.User, error) {
	address = strings.ToLower(address)
	var u models.User
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (address) VALUES ($1)
		ON CONFLICT (address) DO UPDATE SET updated_at = NOW()
		RETURNING id, address, did
	`, address).Scan(&u.ID, &u.Address, &u.DID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) BindDID(ctx context.Context, address, did string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET did = $2, updated_at = NOW() WHERE LOWER(address) = LOWER($1)
	`, address, did)
	return err
}

func (r *UserRepo) GetByAddress(ctx context.Context, address string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, address, did FROM users WHERE LOWER(address) = LOWER($1)`, address,
	).Scan(&u.ID, &u.Address, &u.DID)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

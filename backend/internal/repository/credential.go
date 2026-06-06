// Package repository 可验证凭证仓储
package repository

// 导入依赖
import (
	"context"       // 上下文
	"encoding/json" // JSON
	"time"          // 时间

	"github.com/jackc/pgx/v5/pgxpool" // 连接池
)

// Credential 凭证数据模型
type Credential struct {
	ID             int64           `json:"id"`              // 主键
	UserAddress    string          `json:"user_address"`    // 用户钱包地址
	CredentialType string          `json:"credential_type"` // 凭证类型
	VCJSON         json.RawMessage `json:"vc_json"`         // VC 原始 JSON
	ExpiresAt      time.Time       `json:"expires_at"`      // 过期时间
	Revoked        bool            `json:"revoked"`         // 是否撤销
	CreatedAt      time.Time       `json:"created_at"`      // 创建时间
}

// CredentialRepo 凭证仓储
type CredentialRepo struct {
	pool *pgxpool.Pool // 数据库连接池
}

// NewCredentialRepo 创建凭证仓储
func NewCredentialRepo(pool *pgxpool.Pool) *CredentialRepo {
	return &CredentialRepo{pool: pool}
}

// Insert 插入新凭证，返回生成的 ID
func (r *CredentialRepo) Insert(ctx context.Context, c Credential) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO credentials (user_address, credential_type, vc_json, expires_at)
		VALUES (LOWER($1), $2, $3, $4)
		RETURNING id
	`, c.UserAddress, c.CredentialType, c.VCJSON, c.ExpiresAt).Scan(&id)
	return id, err
}

// ListByUser 查询指定用户所有有效凭证（未撤销且未过期）
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
	// 逐行扫描
	for rows.Next() {
		var c Credential
		if err := rows.Scan(&c.ID, &c.UserAddress, &c.CredentialType, &c.VCJSON, &c.ExpiresAt, &c.Revoked, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// HasValidType 判断用户是否持有某类型的有效凭证
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

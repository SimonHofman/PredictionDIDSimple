// Package repository 提供数据库持久层的仓储实现
package repository

// 导入依赖
import (
	"context" // 上下文，用于控制请求生命周期
	"strings" // 字符串处理工具

	"github.com/jackc/pgx/v5/pgxpool"                  // PostgreSQL 连接池
	"github.com/prediction-did/simple/internal/models" // 数据模型定义
)

// UserRepo 用户仓储结构体，封装用户相关的数据库操作
type UserRepo struct {
	pool *pgxpool.Pool // 数据库连接池实例
}

// NewUserRepo 创建新的用户仓储实例
// 参数 pool: 数据库连接池
// 返回: UserRepo 指针
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool} // 初始化并返回仓储实例
}

// UpsertByAddress 通过钱包地址创建或更新用户
// 如果用户不存在则插入新记录，如果已存在则更新时间戳
// 参数 address: 用户钱包地址
// 返回: 用户对象和错误信息
func (r *UserRepo) UpsertByAddress(ctx context.Context, address string) (*models.User, error) {
	address = strings.ToLower(address) // 将地址转换为小写，确保一致性
	var u models.User                  // 初始化用户对象
	// 执行 upsert 操作：插入用户，如果地址已存在则更新时间戳
	// RETURNING 子句返回操作后的用户数据
	err := r.pool.QueryRow(ctx, `
		INSERT INTO users (address) VALUES ($1)
		ON CONFLICT (address) DO UPDATE SET updated_at = NOW()
		RETURNING id, address, did
	`, address).Scan(&u.ID, &u.Address, &u.DID)
	if err != nil {
		return nil, err // 操作出错时返回错误
	}
	return &u, nil // 返回用户对象
}

// BindDID 为用户绑定去中心化身份标识（DID）
// 参数 address: 用户钱包地址
// 参数 did: 去中心化身份标识字符串
// 返回: 错误信息
func (r *UserRepo) BindDID(ctx context.Context, address, did string) error {
	// 更新用户的 DID 字段，通过地址匹配（不区分大小写）
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET did = $2, updated_at = NOW() WHERE LOWER(address) = LOWER($1)
	`, address, did)
	return err // 返回执行结果的错误信息
}

// GetByAddress 通过钱包地址查询用户信息
// 参数 address: 用户钱包地址
// 返回: 用户对象和错误信息
func (r *UserRepo) GetByAddress(ctx context.Context, address string) (*models.User, error) {
	var u models.User // 初始化用户对象
	// 通过地址查询用户，不区分大小写
	err := r.pool.QueryRow(ctx,
		`SELECT id, address, did FROM users WHERE LOWER(address) = LOWER($1)`, address,
	).Scan(&u.ID, &u.Address, &u.DID)
	if err != nil {
		return nil, err // 查询出错时返回错误
	}
	return &u, nil // 返回用户对象
}

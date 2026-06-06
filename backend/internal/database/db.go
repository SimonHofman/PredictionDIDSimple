// Package database 数据库连接池、Ping 检测、迁移执行
package database

// 导入依赖
import (
	"context" // 上下文
	"fmt"     // 错误格式化
	"time"    // 超时

	"github.com/golang-migrate/migrate/v4"                     // 数据库迁移核心
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres 驱动
	_ "github.com/golang-migrate/migrate/v4/source/file"       // 文件源
	"github.com/jackc/pgx/v5/pgxpool"                          // pgx 连接池
)

// NewPool 创建 PostgreSQL 连接池
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// 使用 pgxpool 创建连接池
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return pool, nil
}

// Ping 用于探活数据库连接（3 秒超时）
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second) // 3 秒超时
	defer cancel()
	return pool.Ping(ctx) // 执行 ping
}

// RunMigrations 执行数据库迁移（向上）
func RunMigrations(databaseURL, migrationsPath string) error {
	// 创建迁移器：source 为文件系统
	m, err := migrate.New(
		"file://"+migrationsPath, // 迁移文件路径
		databaseURL,              // 数据库 URL
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close() // 关闭迁移器

	// 执行向上迁移，无变更不报错
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

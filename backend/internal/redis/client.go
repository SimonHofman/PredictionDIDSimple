// Package redis 封装 Redis 客户端创建与 Ping 检测
package redis

// 导入依赖
import (
	"context" // 上下文
	"time"    // 超时

	"github.com/redis/go-redis/v9" // Redis 客户端库
)

// NewClient 根据 Redis URL 创建客户端实例
func NewClient(redisURL string) (*redis.Client, error) {
	// 解析 URL 为选项
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	// 创建并返回客户端
	return redis.NewClient(opts), nil
}

// Ping 探活 Redis 连接（2 秒超时）
func Ping(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second) // 2 秒超时
	defer cancel()
	return client.Ping(ctx).Err() // 执行 PING 命令
}

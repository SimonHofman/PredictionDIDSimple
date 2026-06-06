// Package main 是数据同步工具入口（双数据源主源同步）
package main

// 导入依赖
import (
	"context"       // 上下文
	"log"           // 日志
	"os"            // 文件操作
	"path/filepath" // 路径处理

	"github.com/prediction-did/simple/internal/config"     // 配置
	"github.com/prediction-did/simple/internal/database"   // 数据库
	"github.com/prediction-did/simple/internal/repository" // 仓储
	"github.com/prediction-did/simple/internal/wcprovider" // 双数据源 provider
)

// main 从主数据源同步比赛数据到数据库
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background() // 根上下文
	// 创建连接池
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close() // 退出关闭

	// 主、备数据文件路径
	primary := cfg.MockMatchesPath
	secondary := cfg.MockMatchesSecondary
	// 主源文件不存在则使用默认路径
	if _, err := os.Stat(primary); os.IsNotExist(err) {
		primary = filepath.Join("backend", "data", "mock_matches.json")
	}
	// 备源文件不存在则使用默认路径
	if _, err := os.Stat(secondary); os.IsNotExist(err) {
		secondary = filepath.Join("backend", "data", "mock_matches_secondary.json")
	}
	// 创建双源 provider
	dual := wcprovider.NewDual(primary, secondary)
	// 主源同步
	n, err := dual.SyncPrimary(ctx, repository.NewMatchRepo(pool))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("synced %d matches (primary source)", n) // 输出同步数量
}

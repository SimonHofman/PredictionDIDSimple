// Package main 是种子数据加载工具的入口
package main

// 导入依赖
import (
	"context"       // 上下文
	"log"           // 日志
	"os"            // 文件操作
	"path/filepath" // 路径处理

	"github.com/prediction-did/simple/internal/config"     // 配置加载
	"github.com/prediction-did/simple/internal/database"   // 数据库
	"github.com/prediction-did/simple/internal/repository" // 仓储层
	"github.com/prediction-did/simple/internal/wcprovider" // mock 数据源
)

// main 加载 mock_matches.json，将比赛数据写入数据库（用于本地测试）
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background() // 根上下文

	// 计算迁移目录
	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("backend", "migrations")
	}
	// 先执行迁移确保表存在
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatal(err)
	}

	// 创建连接池
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close() // 退出关闭

	// 解析 mock 数据文件路径，兼容不同运行目录
	path := cfg.MockMatchesPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join("backend", "data", "mock_matches.json")
	}
	// 同步比赛数据到数据库，返回写入数量
	n, err := wcprovider.NewMock(path).Sync(ctx, repository.NewMatchRepo(pool))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d matches", n) // 输出种子数量
}

// Package main 是数据库迁移工具的入口
package main

// 导入依赖
import (
	"log"           // 日志
	"os"            // 文件操作
	"path/filepath" // 路径拼接

	"github.com/prediction-did/simple/internal/config"   // 配置加载
	"github.com/prediction-did/simple/internal/database" // 数据库迁移工具
)

// main 执行数据库迁移
func main() {
	// 加载配置以获取数据库 URL
	cfg, err := config.Load()
	if err != nil {
		// 配置错误终止
		log.Fatalf("config: %v", err)
	}

	// 计算迁移文件目录
	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// 兼容运行目录不在 backend 下的场景
		migrationsPath = filepath.Join("backend", "migrations")
	}

	// 执行迁移
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		// 迁移失败终止
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations ok") // 成功日志
}

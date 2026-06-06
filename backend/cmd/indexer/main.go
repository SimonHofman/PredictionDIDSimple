// Package main 是独立索引器服务的入口
package main

// 导入所需依赖
import (
	"context"   // 上下文控制
	"log"       // 日志
	"os/signal" // 系统信号
	"syscall"   // 系统调用常量

	"github.com/prediction-did/simple/internal/config"     // 配置加载
	"github.com/prediction-did/simple/internal/database"   // 数据库
	"github.com/prediction-did/simple/internal/indexer"    // 索引器实现
	"github.com/prediction-did/simple/internal/repository" // 仓储层
)

// main 启动索引器进程，监听链上事件并写入数据库
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err) // 失败终止
	}
	// 创建可被信号取消的上下文（SIGINT/SIGTERM）
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel() // 退出时取消

	// 创建数据库连接池
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err) // 数据库连接失败终止
	}
	defer pool.Close() // 退出时关闭连接池

	// 创建索引器实例并注入所需仓储
	idx, err := indexer.New(
		ctx, cfg,
		repository.NewMatchRepo(pool),        // 比赛仓储
		repository.NewMarketRepo(pool),       // 市场仓储
		repository.NewPositionRepo(pool),     // 持仓仓储
		repository.NewIndexerStateRepo(pool), // 索引器状态仓储
	)
	if err != nil {
		log.Fatal(err) // 索引器初始化失败终止
	}
	log.Println("indexer started") // 启动日志
	// 运行索引器，正常取消不视为错误
	if err := idx.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

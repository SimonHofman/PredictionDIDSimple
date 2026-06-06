// Package main 是 Oracle 工作进程的入口
package main

// 导入依赖
import (
	"context"   // 上下文
	"log"       // 日志
	"os/signal" // 信号监听
	"syscall"   // 信号常量

	"github.com/prediction-did/simple/internal/alert"      // 告警通知
	"github.com/prediction-did/simple/internal/blockchain" // 区块链客户端
	"github.com/prediction-did/simple/internal/config"     // 配置
	"github.com/prediction-did/simple/internal/database"   // 数据库
	"github.com/prediction-did/simple/internal/oracle"     // Oracle Worker
	"github.com/prediction-did/simple/internal/repository" // 仓储层
	"github.com/prediction-did/simple/internal/wcprovider" // WC 数据源
)

// main 启动 Oracle 工作者，处理结算任务
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err) // 配置错误终止
	}
	// 信号上下文（支持优雅退出）
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 创建数据库连接池
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err) // 数据库错误终止
	}
	defer pool.Close() // 退出时关闭

	// Oracle 客户端变量
	var chain *blockchain.OracleClient
	// 仅当地址与私钥都配置时初始化链上 Oracle 客户端
	if cfg.OracleAdapterAddress != "" && cfg.OraclePrivateKey != "" {
		chain, err = blockchain.NewOracleClient(ctx, cfg.EthRPCURL, cfg.OracleAdapterAddress, cfg.OraclePrivateKey, cfg.ChainID)
		if err != nil {
			log.Fatal(err) // 创建失败终止
		}
	} else {
		// 提示用户配置环境变量
		log.Println("WARN: oracle chain client disabled (set ORACLE_ADAPTER_ADDRESS + ORACLE_PRIVATE_KEY)")
	}

	// 创建 Oracle Worker，并注入仓储与外部依赖
	worker := oracle.NewWorker(
		cfg,
		repository.NewOracleJobRepo(pool), // Oracle 任务仓储
		repository.NewMatchRepo(pool),     // 比赛仓储
		repository.NewMarketRepo(pool),    // 市场仓储
		wcprovider.NewDual(cfg.MockMatchesPath, cfg.MockMatchesSecondary), // 双源 mock 数据
		chain,                          // 链上 Oracle 客户端
		alert.New(cfg.AlertWebhookURL), // 告警通知器
	)
	log.Println("oracle worker started") // 启动日志
	// 运行 worker 主循环
	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err) // 异常退出
	}
}

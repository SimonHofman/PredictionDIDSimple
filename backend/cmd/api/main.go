// Package main 是 API 服务的入口包
package main

// 导入所需的标准库及第三方依赖
import (
	"context"       // 上下文，用于控制生命周期与取消信号
	"log"           // 日志输出
	"net/http"      // HTTP 服务相关常量
	"os"            // 操作系统功能（文件、信号等）
	"os/signal"     // 系统信号监听
	"path/filepath" // 路径拼接工具
	"syscall"       // 系统调用常量（如 SIGINT/SIGTERM）
	"time"          // 时间相关功能

	"github.com/redis/go-redis/v9" // Redis 客户端 v9 版本

	"github.com/prediction-did/simple/internal/blockchain"        // 区块链客户端封装
	"github.com/prediction-did/simple/internal/config"            // 配置加载
	"github.com/prediction-did/simple/internal/database"          // 数据库连接与迁移
	"github.com/prediction-did/simple/internal/indexer"           // 链上事件索引器
	redisclient "github.com/prediction-did/simple/internal/redis" // Redis 封装客户端
	"github.com/prediction-did/simple/internal/repository"        // 数据仓储层
	"github.com/prediction-did/simple/internal/server"            // HTTP 服务器
)

// main 是 API 程序的入口函数
func main() {
	// 加载配置（环境变量或配置文件）
	cfg, err := config.Load()
	if err != nil {
		// 配置加载失败，直接终止程序
		log.Fatalf("config: %v", err)
	}

	// 创建可取消的上下文，便于优雅关闭子任务
	ctx, cancel := context.WithCancel(context.Background())
	// 程序退出时取消上下文
	defer cancel()

	// 计算 migrations 目录路径，先尝试当前目录
	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// 如果不存在，尝试 backend/migrations 路径
		migrationsPath = filepath.Join("backend", "migrations")
	}
	// 执行数据库迁移
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		// 迁移失败立即终止
		log.Fatalf("migrate: %v", err)
	}
	// 输出迁移成功日志
	log.Println("migrations applied")

	// 创建数据库连接池
	dbPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		// 数据库连接失败终止
		log.Fatalf("database: %v", err)
	}
	// 程序退出时关闭连接池
	defer dbPool.Close()

	// Redis 客户端变量声明
	var rdb *redis.Client
	// 创建 Redis 客户端
	redisClient, err := redisclient.NewClient(cfg.RedisURL)
	if err != nil {
		// Redis 解析失败仅警告，不终止程序
		log.Printf("WARN: redis parse: %v", err)
	} else {
		// 保存可用的 Redis 客户端
		rdb = redisClient
		// 通过 Ping 检测连通性
		if err := redisclient.Ping(ctx, redisClient); err != nil {
			// 连通性失败时降级运行
			log.Printf("WARN: redis ping: %v (degraded)", err)
		} else {
			// 连接成功日志
			log.Println("redis connected")
		}
	}

	// 创建区块链客户端
	chain := blockchain.New(cfg.EthRPCURL, cfg.ChainID)
	// 启动后台 ping 检测链上节点状态
	chain.StartBackgroundPing(ctx)

	// 如果配置了 Factory 地址，则启动索引器
	if cfg.FactoryAddress != "" {
		go func() {
			// 创建索引器实例，注入所需仓储
			idx, err := indexer.New(
				ctx, cfg,
				repository.NewMatchRepo(dbPool),        // 比赛仓储
				repository.NewMarketRepo(dbPool),       // 市场仓储
				repository.NewPositionRepo(dbPool),     // 持仓仓储
				repository.NewIndexerStateRepo(dbPool), // 索引器状态仓储
			)
			if err != nil {
				// 初始化失败仅打印
				log.Printf("indexer init: %v", err)
				return
			}
			// 运行索引器主循环
			if err := idx.Run(ctx); err != nil && err != context.Canceled {
				// 异常打印
				log.Printf("indexer: %v", err)
			}
		}()
	}

	// Oracle 客户端变量声明
	var oracleChain *blockchain.OracleClient
	// 仅当 Oracle 地址与私钥都配置时才创建管理员客户端
	if cfg.OracleAdapterAddress != "" && cfg.OraclePrivateKey != "" {
		oracleChain, err = blockchain.NewOracleClient(ctx, cfg.EthRPCURL, cfg.OracleAdapterAddress, cfg.OraclePrivateKey, cfg.ChainID)
		if err != nil {
			// Oracle 客户端创建失败仅警告
			log.Printf("WARN: oracle admin client: %v", err)
		}
	}

	// 创建 HTTP 服务器，注入所有依赖
	srv := server.New(server.Dependencies{
		Port:        cfg.HTTPPort, // 监听端口
		Cfg:         cfg,          // 全局配置
		DB:          dbPool,       // 数据库连接池
		Redis:       rdb,          // Redis 客户端
		Chain:       chain,        // 区块链只读客户端
		OracleChain: oracleChain,  // Oracle 写入客户端
	})

	// 在协程中启动 HTTP 服务器，避免阻塞
	go func() {
		// 打印监听地址
		log.Printf("API listening on %s", srv.String())
		// 启动监听，如果非正常关闭则致命错误
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	// 创建信号通道用于优雅关闭
	quit := make(chan os.Signal, 1)
	// 监听 Ctrl+C 与 kill 命令
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞直到收到信号
	<-quit
	// 输出关闭日志
	log.Println("shutting down...")

	// 创建带 10 秒超时的关闭上下文
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	// 优雅关闭服务器
	if err := srv.Shutdown(shutdownCtx); err != nil {
		// 关闭失败打印警告
		log.Printf("shutdown: %v", err)
	}
}

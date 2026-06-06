// Package main 是对账（reconcile）工具的入口
package main

// 导入依赖
import (
	"context"       // 上下文
	"log"           // 日志
	"math/big"      // 大整数运算
	"os"            // 文件操作
	"path/filepath" // 路径处理

	"github.com/jackc/pgx/v5/pgxpool"                      // PostgreSQL 连接池
	"github.com/prediction-did/simple/internal/blockchain" // 区块链工具
	"github.com/prediction-did/simple/internal/config"     // 配置加载
	"github.com/prediction-did/simple/internal/database"   // 数据库工具
)

// main 启动对账任务，校验链上余额与数据库总额是否一致
func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err) // 配置失败终止
	}
	ctx := context.Background() // 使用根上下文
	// 计算迁移目录路径
	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// 兼容路径
		migrationsPath = filepath.Join("backend", "migrations")
	}
	// 先执行迁移，确保表结构最新
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	// 创建数据库连接池
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close() // 退出关闭

	// 校验抵押代币地址是否已配置
	if cfg.CollateralAddress == "" {
		log.Fatal("MOCK_USDC_ADDRESS required for reconcile")
	}
	// 执行对账逻辑
	if err := run(ctx, pool, cfg); err != nil {
		log.Fatalf("reconcile: %v", err)
	}
}

// run 是对账主流程：读取数据库中的市场总额，对比链上余额并写入对账记录
func run(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	// 查询所有 OPEN/RESOLVED 状态的市场及其数据库累计
	rows, err := pool.Query(ctx, `
		SELECT market_address, COALESCE(yes_pool,0)+COALESCE(no_pool,0) AS db_total
		FROM markets WHERE status IN ('OPEN','RESOLVED')
	`)
	if err != nil {
		return err
	}
	defer rows.Close() // 退出关闭游标

	rpc := cfg.EthRPCURL // 主 RPC 节点
	// 遍历每个市场
	for rows.Next() {
		var addr, dbTotal string // 市场地址、数据库累计
		if err := rows.Scan(&addr, &dbTotal); err != nil {
			return err
		}
		// 主 RPC 查询链上余额
		chainBal, err := blockchain.ERC20Balance(ctx, rpc, cfg.CollateralAddress, addr)
		// 主 RPC 失败则回退到备用 RPC
		if err != nil && cfg.EthRPCFallback != "" {
			chainBal, err = blockchain.ERC20Balance(ctx, cfg.EthRPCFallback, cfg.CollateralAddress, addr)
		}
		if err != nil {
			// 仍失败则跳过该市场
			log.Printf("WARN %s: %v", addr, err)
			continue
		}
		db := new(big.Int)                      // 数据库累计大整数
		db.SetString(dbTotal, 10)               // 解析十进制字符串
		delta := new(big.Int).Sub(chainBal, db) // 差值 = 链上 - 数据库
		ok := delta.Cmp(big.NewInt(0)) == 0     // 差值为 0 即一致
		// 写入对账记录表
		_, _ = pool.Exec(ctx, `
			INSERT INTO reconciliation_runs (market_address, db_total, chain_balance, delta, ok)
			VALUES ($1,$2::numeric,$3::numeric,$4::numeric,$5)
		`, addr, dbTotal, chainBal.String(), delta.String(), ok)
		// 输出对账结果日志
		log.Printf("market=%s db=%s chain=%s ok=%v", addr, dbTotal, chainBal.String(), ok)
	}
	return rows.Err() // 返回行扫描错误（若有）
}

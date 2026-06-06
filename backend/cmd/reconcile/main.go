package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()
	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("backend", "migrations")
	}
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if cfg.CollateralAddress == "" {
		log.Fatal("MOCK_USDC_ADDRESS required for reconcile")
	}
	if err := run(ctx, pool, cfg); err != nil {
		log.Fatalf("reconcile: %v", err)
	}
}

func run(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	rows, err := pool.Query(ctx, `
		SELECT market_address, COALESCE(yes_pool,0)+COALESCE(no_pool,0) AS db_total
		FROM markets WHERE status IN ('OPEN','RESOLVED')
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	rpc := cfg.EthRPCURL
	for rows.Next() {
		var addr, dbTotal string
		if err := rows.Scan(&addr, &dbTotal); err != nil {
			return err
		}
		chainBal, err := blockchain.ERC20Balance(ctx, rpc, cfg.CollateralAddress, addr)
		if err != nil && cfg.EthRPCFallback != "" {
			chainBal, err = blockchain.ERC20Balance(ctx, cfg.EthRPCFallback, cfg.CollateralAddress, addr)
		}
		if err != nil {
			log.Printf("WARN %s: %v", addr, err)
			continue
		}
		db := new(big.Int)
		db.SetString(dbTotal, 10)
		delta := new(big.Int).Sub(chainBal, db)
		ok := delta.Cmp(big.NewInt(0)) == 0
		_, _ = pool.Exec(ctx, `
			INSERT INTO reconciliation_runs (market_address, db_total, chain_balance, delta, ok)
			VALUES ($1,$2::numeric,$3::numeric,$4::numeric,$5)
		`, addr, dbTotal, chainBal.String(), delta.String(), ok)
		log.Printf("market=%s db=%s chain=%s ok=%v", addr, dbTotal, chainBal.String(), ok)
	}
	return rows.Err()
}

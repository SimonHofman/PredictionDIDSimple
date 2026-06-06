package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
	"github.com/prediction-did/simple/internal/indexer"
	redisclient "github.com/prediction-did/simple/internal/redis"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("backend", "migrations")
	}
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied")

	dbPool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer dbPool.Close()

	var rdb *redis.Client
	redisClient, err := redisclient.NewClient(cfg.RedisURL)
	if err != nil {
		log.Printf("WARN: redis parse: %v", err)
	} else {
		rdb = redisClient
		if err := redisclient.Ping(ctx, redisClient); err != nil {
			log.Printf("WARN: redis ping: %v (degraded)", err)
		} else {
			log.Println("redis connected")
		}
	}

	chain := blockchain.New(cfg.EthRPCURL, cfg.ChainID)
	chain.StartBackgroundPing(ctx)

	if cfg.FactoryAddress != "" {
		go func() {
			idx, err := indexer.New(
				ctx, cfg,
				repository.NewMatchRepo(dbPool),
				repository.NewMarketRepo(dbPool),
				repository.NewPositionRepo(dbPool),
				repository.NewIndexerStateRepo(dbPool),
			)
			if err != nil {
				log.Printf("indexer init: %v", err)
				return
			}
			if err := idx.Run(ctx); err != nil && err != context.Canceled {
				log.Printf("indexer: %v", err)
			}
		}()
	}

	var oracleChain *blockchain.OracleClient
	if cfg.OracleAdapterAddress != "" && cfg.OraclePrivateKey != "" {
		oracleChain, err = blockchain.NewOracleClient(ctx, cfg.EthRPCURL, cfg.OracleAdapterAddress, cfg.OraclePrivateKey, cfg.ChainID)
		if err != nil {
			log.Printf("WARN: oracle admin client: %v", err)
		}
	}

	srv := server.New(server.Dependencies{
		Port:        cfg.HTTPPort,
		Cfg:         cfg,
		DB:          dbPool,
		Redis:       rdb,
		Chain:       chain,
		OracleChain: oracleChain,
	})

	go func() {
		log.Printf("API listening on %s", srv.String())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

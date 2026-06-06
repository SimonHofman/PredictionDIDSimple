package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/prediction-did/simple/internal/alert"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
	"github.com/prediction-did/simple/internal/oracle"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/wcprovider"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var chain *blockchain.OracleClient
	if cfg.OracleAdapterAddress != "" && cfg.OraclePrivateKey != "" {
		chain, err = blockchain.NewOracleClient(ctx, cfg.EthRPCURL, cfg.OracleAdapterAddress, cfg.OraclePrivateKey, cfg.ChainID)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("WARN: oracle chain client disabled (set ORACLE_ADAPTER_ADDRESS + ORACLE_PRIVATE_KEY)")
	}

	worker := oracle.NewWorker(
		cfg,
		repository.NewOracleJobRepo(pool),
		repository.NewMatchRepo(pool),
		repository.NewMarketRepo(pool),
		wcprovider.NewDual(cfg.MockMatchesPath, cfg.MockMatchesSecondary),
		chain,
		alert.New(cfg.AlertWebhookURL),
	)
	log.Println("oracle worker started")
	if err := worker.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
	"github.com/prediction-did/simple/internal/indexer"
	"github.com/prediction-did/simple/internal/repository"
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

	idx, err := indexer.New(
		ctx, cfg,
		repository.NewMatchRepo(pool),
		repository.NewMarketRepo(pool),
		repository.NewPositionRepo(pool),
		repository.NewIndexerStateRepo(pool),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("indexer started")
	if err := idx.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}

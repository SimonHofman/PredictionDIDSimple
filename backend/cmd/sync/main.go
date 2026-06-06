package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/wcprovider"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	primary := cfg.MockMatchesPath
	secondary := cfg.MockMatchesSecondary
	if _, err := os.Stat(primary); os.IsNotExist(err) {
		primary = filepath.Join("backend", "data", "mock_matches.json")
	}
	if _, err := os.Stat(secondary); os.IsNotExist(err) {
		secondary = filepath.Join("backend", "data", "mock_matches_secondary.json")
	}
	dual := wcprovider.NewDual(primary, secondary)
	n, err := dual.SyncPrimary(ctx, repository.NewMatchRepo(pool))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("synced %d matches (primary source)", n)
}

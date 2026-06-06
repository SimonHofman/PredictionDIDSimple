package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
	"github.com/prediction-did/simple/internal/wcprovider"
	"github.com/prediction-did/simple/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("backend", "migrations")
	}
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatal(err)
	}

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	path := cfg.MockMatchesPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join("backend", "data", "mock_matches.json")
	}
	n, err := wcprovider.NewMock(path).Sync(ctx, repository.NewMatchRepo(pool))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d matches", n)
}

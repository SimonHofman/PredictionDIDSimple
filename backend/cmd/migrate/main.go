package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	migrationsPath := filepath.Join("migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = filepath.Join("backend", "migrations")
	}

	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations ok")
}

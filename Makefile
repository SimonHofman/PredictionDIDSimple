.PHONY: up down migrate seed sync-matches dev-api dev-oracle dev-web test-contracts test-backend test-frontend sync-abi build-all

up:
	docker compose up -d

down:
	docker compose down

migrate:
	cd backend && go run ./cmd/migrate

seed:
	cd backend && go run ./cmd/seed

sync-matches:
	cd backend && go run ./cmd/sync

dev-api:
	cd backend && go run ./cmd/api

dev-oracle:
	cd backend && go run ./cmd/oracle

dev-web:
	cd frontend && npm run dev

test-contracts:
	cd contracts && npm test

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm run build

sync-abi:
	cd contracts && npm run compile && npm run export-abi

build-all: sync-abi test-backend test-frontend

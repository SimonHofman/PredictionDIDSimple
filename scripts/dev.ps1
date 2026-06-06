# Phase 0 local dev helper (Windows PowerShell)
# Usage: .\scripts\dev.ps1 up | migrate | api | web | contracts | sync-abi
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("up", "down", "migrate", "api", "web", "contracts", "sync-abi", "test")]
    [string]$Action
)

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

switch ($Action) {
    "up" { docker compose up -d }
    "down" { docker compose down }
    "migrate" {
        $env:DATABASE_URL = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable" }
        Set-Location backend
        go run ./cmd/migrate
    }
    "api" {
        $env:DATABASE_URL = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable" }
        Set-Location backend
        go run ./cmd/api
    }
    "web" {
        Set-Location frontend
        npm run dev
    }
    "contracts" {
        Set-Location contracts
        npx hardhat node
    }
    "sync-abi" {
        & "$Root\scripts\sync-abi.ps1"
    }
    "test" {
        Set-Location contracts
        npm test
        Set-Location $Root\backend
        go test ./...
        Set-Location $Root\frontend
        npm run build
    }
}

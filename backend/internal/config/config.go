package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort              string
	DatabaseURL           string
	RedisURL              string
	EthRPCURL             string
	EthRPCFallback        string
	ChainID               int64
	JWTSecret             string
	SIWEDomain            string
	SIWEURI               string
	FactoryAddress        string
	FactoryV3Address      string
	CollateralAddress     string
	OracleAdapterAddress  string
	OracleAdapterV2       string
	DIDRegistryAddress    string
	OraclePrivateKey      string
	IndexerStartBlock     uint64
	IndexerPollSeconds    int
	MockMatchesPath       string
	MockMatchesSecondary  string
	OracleCooldownMinutes int
	OracleTimelockSeconds int
	OraclePollSeconds     int
	AdminAPIKey           string
	VCIssuerKey           string
	AlertWebhookURL       string
	BlockedCountries      map[string]bool
	GeoBlockEnabled       bool
	RateLimitPerMinute    int
	KYCWebhookSecret      string
	ComplianceRequired    bool
	Environment           string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:              getEnv("HTTP_PORT", "8080"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EthRPCURL:             getEnv("ETH_RPC_URL", "http://127.0.0.1:8545"),
		EthRPCFallback:        os.Getenv("ETH_RPC_FALLBACK_URL"),
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		SIWEDomain:            getEnv("SIWE_DOMAIN", "localhost"),
		SIWEURI:               getEnv("SIWE_URI", "http://localhost:5173"),
		FactoryAddress:        os.Getenv("MARKET_FACTORY_ADDRESS"),
		FactoryV3Address:      os.Getenv("MARKET_FACTORY_V3_ADDRESS"),
		CollateralAddress:     os.Getenv("MOCK_USDC_ADDRESS"),
		OracleAdapterAddress:  os.Getenv("ORACLE_ADAPTER_ADDRESS"),
		OracleAdapterV2:       os.Getenv("ORACLE_ADAPTER_V2_ADDRESS"),
		DIDRegistryAddress:    os.Getenv("DID_REGISTRY_ADDRESS"),
		OraclePrivateKey:      os.Getenv("ORACLE_PRIVATE_KEY"),
		IndexerPollSeconds:    getEnvInt("INDEXER_POLL_SECONDS", 5),
		MockMatchesPath:       getEnv("MOCK_MATCHES_PATH", "data/mock_matches.json"),
		MockMatchesSecondary:  getEnv("MOCK_MATCHES_SECONDARY_PATH", "data/mock_matches_secondary.json"),
		OracleCooldownMinutes: getEnvInt("ORACLE_COOLDOWN_MINUTES", 15),
		OracleTimelockSeconds: getEnvInt("ORACLE_TIMELOCK_SECONDS", 120),
		OraclePollSeconds:     getEnvInt("ORACLE_POLL_SECONDS", 10),
		AdminAPIKey:           os.Getenv("ADMIN_API_KEY"),
		VCIssuerKey:           getEnv("VC_ISSUER_KEY", "dev-vc-issuer-key"),
		AlertWebhookURL:       os.Getenv("ALERT_WEBHOOK_URL"),
		GeoBlockEnabled:       getEnv("GEO_BLOCK_ENABLED", "true") == "true",
		RateLimitPerMinute:    getEnvInt("RATE_LIMIT_PER_MINUTE", 120),
		KYCWebhookSecret:      os.Getenv("KYC_WEBHOOK_SECRET"),
		ComplianceRequired:    getEnv("COMPLIANCE_REQUIRED", "true") == "true",
		Environment:           getEnv("APP_ENV", "development"),
	}
	cfg.BlockedCountries = parseBlockedCountries(getEnv("BLOCKED_COUNTRIES", "US,KP,IR,SY"))

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (see .env.example)")
	}

	chainIDStr := getEnv("CHAIN_ID", "31337")
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid CHAIN_ID %q: %w", chainIDStr, err)
	}
	cfg.ChainID = chainID

	startBlock, err := strconv.ParseUint(getEnv("INDEXER_START_BLOCK", "0"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid INDEXER_START_BLOCK: %w", err)
	}
	cfg.IndexerStartBlock = startBlock

	return cfg, nil
}

func parseBlockedCountries(raw string) map[string]bool {
	m := make(map[string]bool)
	for _, part := range strings.Split(raw, ",") {
		c := strings.TrimSpace(strings.ToUpper(part))
		if c != "" {
			m[c] = true
		}
	}
	return m
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// Package config 加载并解析运行时配置（环境变量）
package config

// 导入依赖
import (
	"fmt"     // 错误格式化
	"os"      // 环境变量
	"strconv" // 字符串转数字
	"strings" // 字符串处理
)

// Config 应用全局配置结构体
type Config struct {
	HTTPPort              string          // HTTP 监听端口
	DatabaseURL           string          // PostgreSQL 连接字符串
	RedisURL              string          // Redis 连接地址
	EthRPCURL             string          // 以太坊主 RPC 地址
	EthRPCFallback        string          // 以太坊备用 RPC
	ChainID               int64           // 链 ID
	JWTSecret             string          // JWT 签名密钥
	SIWEDomain            string          // SIWE 校验域名
	SIWEURI               string          // SIWE 校验 URI
	FactoryAddress        string          // MarketFactory 合约地址
	FactoryV3Address      string          // MarketFactory V3 合约地址
	CollateralAddress     string          // 抵押代币（如 USDC）合约地址
	OracleAdapterAddress  string          // OracleAdapter 合约地址
	OracleAdapterV2       string          // OracleAdapter V2 合约地址
	DIDRegistryAddress    string          // DID Registry 合约地址
	OraclePrivateKey      string          // Oracle 签名私钥
	IndexerStartBlock     uint64          // 索引器起始区块
	IndexerPollSeconds    int             // 索引器轮询间隔（秒）
	MockMatchesPath       string          // mock 比赛数据主源路径
	MockMatchesSecondary  string          // mock 比赛数据备源路径
	OracleCooldownMinutes int             // Oracle 冷却时间（分钟）
	OracleTimelockSeconds int             // Oracle 时间锁（秒）
	OraclePollSeconds     int             // Oracle 轮询间隔（秒）
	AdminAPIKey           string          // 管理员 API 密钥
	VCIssuerKey           string          // 可验证凭证颁发者密钥
	AlertWebhookURL       string          // 告警 Webhook URL
	BlockedCountries      map[string]bool // 封禁国家列表
	GeoBlockEnabled       bool            // 是否开启地理封锁
	RateLimitPerMinute    int             // 每分钟速率限制
	KYCWebhookSecret      string          // KYC 回调密钥
	ComplianceRequired    bool            // 是否需要合规审查
	Environment           string          // 运行环境（development/production）
}

// Load 从环境变量加载配置，缺少必要项时返回错误
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:              getEnv("HTTP_PORT", "8080"),                                               // 默认端口 8080
		DatabaseURL:           os.Getenv("DATABASE_URL"),                                                 // 必填
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),                           // 默认本地 Redis
		EthRPCURL:             getEnv("ETH_RPC_URL", "http://127.0.0.1:8545"),                            // 默认本地节点
		EthRPCFallback:        os.Getenv("ETH_RPC_FALLBACK_URL"),                                         // 备用 RPC
		JWTSecret:             getEnv("JWT_SECRET", "dev-secret-change-in-production"),                   // JWT 密钥
		SIWEDomain:            getEnv("SIWE_DOMAIN", "localhost"),                                        // SIWE 域名
		SIWEURI:               getEnv("SIWE_URI", "http://localhost:5173"),                               // SIWE URI
		FactoryAddress:        os.Getenv("MARKET_FACTORY_ADDRESS"),                                       // Factory 地址
		FactoryV3Address:      os.Getenv("MARKET_FACTORY_V3_ADDRESS"),                                    // Factory V3 地址
		CollateralAddress:     os.Getenv("MOCK_USDC_ADDRESS"),                                            // 抵押代币地址
		OracleAdapterAddress:  os.Getenv("ORACLE_ADAPTER_ADDRESS"),                                       // Oracle 地址
		OracleAdapterV2:       os.Getenv("ORACLE_ADAPTER_V2_ADDRESS"),                                    // Oracle V2 地址
		DIDRegistryAddress:    os.Getenv("DID_REGISTRY_ADDRESS"),                                         // DID 注册地址
		OraclePrivateKey:      os.Getenv("ORACLE_PRIVATE_KEY"),                                           // Oracle 私钥
		IndexerPollSeconds:    getEnvInt("INDEXER_POLL_SECONDS", 5),                                      // 索引轮询 5s
		MockMatchesPath:       getEnv("MOCK_MATCHES_PATH", "data/mock_matches.json"),                     // 主 mock 文件
		MockMatchesSecondary:  getEnv("MOCK_MATCHES_SECONDARY_PATH", "data/mock_matches_secondary.json"), // 备 mock
		OracleCooldownMinutes: getEnvInt("ORACLE_COOLDOWN_MINUTES", 15),                                  // 冷却 15min
		OracleTimelockSeconds: getEnvInt("ORACLE_TIMELOCK_SECONDS", 120),                                 // 时间锁 120s
		OraclePollSeconds:     getEnvInt("ORACLE_POLL_SECONDS", 10),                                      // 轮询 10s
		AdminAPIKey:           os.Getenv("ADMIN_API_KEY"),                                                // 管理员密钥
		VCIssuerKey:           getEnv("VC_ISSUER_KEY", "dev-vc-issuer-key"),                              // VC 发行密钥
		AlertWebhookURL:       os.Getenv("ALERT_WEBHOOK_URL"),                                            // 告警 URL
		GeoBlockEnabled:       getEnv("GEO_BLOCK_ENABLED", "true") == "true",                             // 地理封锁开关
		RateLimitPerMinute:    getEnvInt("RATE_LIMIT_PER_MINUTE", 120),                                   // 每分钟 120 次
		KYCWebhookSecret:      os.Getenv("KYC_WEBHOOK_SECRET"),                                           // KYC 密钥
		ComplianceRequired:    getEnv("COMPLIANCE_REQUIRED", "true") == "true",                           // 合规开关
		Environment:           getEnv("APP_ENV", "development"),                                          // 环境
	}
	// 解析封禁国家列表
	cfg.BlockedCountries = parseBlockedCountries(getEnv("BLOCKED_COUNTRIES", "US,KP,IR,SY"))

	// DATABASE_URL 是必须的
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (see .env.example)")
	}

	// 解析 CHAIN_ID
	chainIDStr := getEnv("CHAIN_ID", "31337")
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid CHAIN_ID %q: %w", chainIDStr, err)
	}
	cfg.ChainID = chainID

	// 解析起始区块号
	startBlock, err := strconv.ParseUint(getEnv("INDEXER_START_BLOCK", "0"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid INDEXER_START_BLOCK: %w", err)
	}
	cfg.IndexerStartBlock = startBlock

	return cfg, nil
}

// parseBlockedCountries 解析逗号分隔的国家代码为 map
func parseBlockedCountries(raw string) map[string]bool {
	m := make(map[string]bool) // 初始化 map
	for _, part := range strings.Split(raw, ",") {
		c := strings.TrimSpace(strings.ToUpper(part)) // 统一大写
		if c != "" {
			m[c] = true // 加入黑名单
		}
	}
	return m
}

// getEnv 获取环境变量，空时返回默认值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt 获取环境变量并转为整数，失败返回默认值
func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key) // 读取
	if v == "" {
		return fallback // 空返回默认
	}
	n, err := strconv.Atoi(v) // 转整数
	if err != nil {
		return fallback // 转换失败返回默认
	}
	return n
}

# Backend 结构体与函数说明文档

本文档按模块整理 `backend/` 目录下 Go 源码中的结构体与函数。Go 无类继承，下文「组合/依赖关系」指结构体嵌入、接口实现或依赖注入。不含 `*_test.go`。

---

## 1. cmd/api（HTTP API 主服务入口）

#### 模块说明

- 可执行入口：启动 HTTP API 服务及可选内嵌 Indexer。

#### 规范与约定

- 失败用 `log.Fatalf` 终止；Redis 不可用仅 WARN 降级。

- 迁移目录优先 `migrations/`，否则 `backend/migrations/`。

#### 组合/依赖关系

- 依赖 config、database、redis、blockchain、indexer、repository、server。

- 本包无自定义结构体。



### 1.1 main

- **函数用途**

  - HTTP API 主进程入口：加载配置、迁移、连接 DB/Redis、启动 RPC 探测、可选 Indexer goroutine、可选 Oracle 客户端、HTTP 服务与优雅关闭。

- **函数参数说明**

  - 无参数（程序入口）。

- **函数返回参数说明**

  - 无返回值；致命错误时进程退出。

- **函数内校验**

  - `config.Load()` 必须成功。

  - PostgreSQL 连接必须成功。

  - 迁移必须成功。

- **函数实现效果**

  - Load → RunMigrations → NewPool → 可选 Redis Ping → blockchain.New + StartBackgroundPing

  - FactoryAddress 非空时 goroutine 启动 indexer.Run

  - Oracle 配置齐全时 NewOracleClient

  - server.New → ListenAndServe → 等待 SIGINT/SIGTERM → Shutdown(10s)

- **错误返回**

  - config/migrate/database 失败：log.Fatalf

  - Redis 失败：WARN 继续

  - Listen 异常：log.Fatalf

---

## 2. cmd/indexer（独立链上索引器）

#### 模块说明

- 独立进程，仅运行链上事件 Indexer。

#### 规范与约定

- 初始化失败 log.Fatal；context.Canceled 正常退出。

#### 组合/依赖关系

- 依赖 config、database、indexer、repository。



### 2.1 main

- **函数用途**

  - 加载配置、连接 DB、构造 Indexer 并阻塞运行直到取消。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；错误 log.Fatal。

- **函数内校验**

  - config.Load 与 database.NewPool 必须成功。

- **函数实现效果**

  - indexer.New → idx.Run(ctx)

- **错误返回**

  - init 失败：log.Fatal

  - Run 非 Cancel 错误：log.Printf

---

## 3. cmd/oracle（Oracle 结算 Worker）

#### 模块说明

- 独立进程，运行 oracle.Worker 自动结算。

#### 规范与约定

- Oracle 链客户端缺失时 WARN，Worker 仍启动（chain=nil 时跳过链上调用）。

#### 组合/依赖关系

- 依赖 config、database、oracle、wcprovider、blockchain、alert、repository。



### 3.1 main

- **函数用途**

  - 初始化 DualProvider、OracleClient（可选）、Worker 并 Run。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；非 Cancel 错误 log.Fatal。

- **函数内校验**

  - config.Load、DB 连接必须成功。

- **函数实现效果**

  - 构造 OracleJobRepo/MatchRepo/MarketRepo → NewDual → NewWorker → Run

- **错误返回**

  - init 失败：log.Fatal

---

## 4. cmd/sync（主数据源赛程同步）

#### 模块说明

- 一次性任务：从 primary mock JSON 同步比赛到 DB。

#### 规范与约定

- 路径 fallback：`backend/data/mock_matches.json` 等。

#### 组合/依赖关系

- 依赖 config、database、wcprovider。



### 4.1 main

- **函数用途**

  - DualProvider.SyncPrimary 写入 matches 表。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；失败 log.Fatal，成功打印条数。

- **函数内校验**

  - config 与 DB 必须可用。

- **函数实现效果**

  - NewDual → SyncPrimary

- **错误返回**

  - sync 失败：log.Fatal

---

## 5. cmd/reconcile（链上/DB 资金对账）

#### 模块说明

- 一次性对账：比较 markets 池子总和与链上 ERC20 余额。

#### 规范与约定

- 需配置 CollateralAddress（MOCK_USDC_ADDRESS）。

#### 组合/依赖关系

- 依赖 config、database、blockchain/erc20。



### 5.1 main

- **函数用途**

  - 加载配置、迁移、执行 run。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；CollateralAddress 空或 run 失败 fatal。

- **函数内校验**

  - CollateralAddress 必填。

- **函数实现效果**

  - RunMigrations → NewPool → run

- **错误返回**

  - 缺少抵押地址：log.Fatal

  - run 失败：log.Fatal

---

### 5.2 run

- **函数用途**

  - 遍历 OPEN/RESOLVED 市场，读链上 balanceOf，写入 reconciliation_runs。

- **函数参数说明**

  - `ctx`（context.Context）：上下文

  - `pool`（*pgxpool.Pool）：DB 连接池

  - `cfg`（*config.Config）：含 RPC 与代币地址

- **函数返回参数说明**

  - `error`：查询/扫描/迭代失败时返回

- **函数内校验**

  - 无业务字段校验；逐行处理市场。

- **函数实现效果**

  - ERC20Balance 失败尝试 fallback RPC；单市场错误 WARN continue；写入对账记录

- **错误返回**

  - SQL 或 rows 错误返回 error

---

## 6. cmd/migrate（数据库迁移）

#### 模块说明

- 仅执行 SQL 迁移 Up。

#### 规范与约定

- 迁移路径同 api。

#### 组合/依赖关系

- 依赖 config、database。



### 6.1 main

- **函数用途**

  - RunMigrations 后退出。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；失败 fatal。

- **函数内校验**

  - DATABASE_URL 有效。

- **函数实现效果**

  - Load → RunMigrations

- **错误返回**

  - migrate 失败：log.Fatalf

---

## 7. cmd/seed（种子数据）

#### 模块说明

- 迁移后从 mock JSON 导入比赛。

#### 规范与约定

- 路径 fallback 同 sync。

#### 组合/依赖关系

- 依赖 config、database、wcprovider/mock。



### 7.1 main

- **函数用途**

  - MockProvider.Sync 导入数据。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值；失败 log.Fatal。

- **函数内校验**

  - config 与 DB 必须成功。

- **函数实现效果**

  - RunMigrations → NewPool → NewMock → Sync

- **错误返回**

  - 失败：log.Fatal

---

## 8. internal/models（API 领域模型）

#### 模块说明

- 定义 REST API 序列化用的核心实体。

#### 规范与约定

- 字段带 `json` tag；无 db tag（repository 层显式 Scan）。

- 金额类字段用 string 存 numeric。

#### 组合/依赖关系

- 被 handler、repository、indexer、wcprovider 引用。



### 8.1 Match 结构体

#### 结构体说明

- 世界杯比赛实体，对应 matches 表/API 响应。

#### 规范与约定

- Status 如 SCHEDULED/LIVE/FINISHED/ORACLE_PENDING/RESOLVED/CANCELLED。

- 比分为可空指针。

#### 组合/依赖关系

- 被 Market.Match 嵌套引用。

#### 结构体字段

- `ID`（int64，json:`id`）：数据库主键

- `ExternalID`（string，json:`external_id`）：外部唯一 ID，如 wc-2026-xxx

- `HomeTeam`（string，json:`home_team`）：主队名称

- `AwayTeam`（string，json:`away_team`）：客队名称

- `KickoffAt`（time.Time，json:`kickoff_at`）：开球时间

- `Status`（string，json:`status`）：比赛状态

- `HomeScore`（*int，json:`home_score,omitempty`）：主队得分，未开赛可为 nil

- `AwayScore`（*int，json:`away_score,omitempty`）：客队得分，未开赛可为 nil

---

### 8.2 Market 结构体

#### 结构体说明

- 预测市场实体，链上市场与 DB 记录的聚合视图。

#### 规范与约定

- YesPool/NoPool 为 parimutuel 池；Reserve* 为 CPMM 储备。

- RequiresVC 控制是否需 VerifiedFan 凭证。

#### 组合/依赖关系

- 可嵌套 *Match；与 Position.Market 关联。

#### 结构体字段

- `ID`（int64）：市场主键

- `MatchID`（*int64）：关联比赛 ID

- `ChainID`（int64）：链 ID

- `FactoryAddress`（string）：工厂合约地址

- `MarketAddress`（string）：市场合约地址

- `OnChainMarketID`（int64）：链上市场序号

- `MatchRef`（string）：链上比赛引用哈希

- `Question`（string）：市场问题描述

- `EndTime`（time.Time）：投注截止时间

- `Status`（string）：OPEN/RESOLVED/VOID 等

- `WinningOutcome`（*int）：结算后获胜 outcome

- `YesPool`（string）：YES 侧池 numeric 字符串

- `NoPool`（string）：NO 侧池 numeric 字符串

- `MarketType`（string）：市场类型标识

- `OutcomeCount`（int）：结果数量

- `FeeBps`（int）：手续费基点

- `ReserveYes`（string）：CPMM YES 储备

- `ReserveNo`（string）：CPMM NO 储备

- `PriceYesBps`（string）：YES 隐含价格基点

- `RequiresVC`（bool）：是否要求 VC 才能交易

- `RestrictedRegion`（string）：限制地区码

- `ResolutionRule`（string）：结算规则 HOME_WIN/OVER_25 等

- `Match`（*Match）：嵌套关联比赛

---

### 8.3 Position 结构体

#### 结构体说明

- 用户在某市场的持仓汇总。

#### 规范与约定

- YesAmount/NoAmount 为累计投注 numeric 字符串。

- Claimed 表示是否已链上领取。

#### 组合/依赖关系

- ListByUser 时 JOIN Market。

#### 结构体字段

- `ID`（int64）：持仓记录主键

- `MarketID`（int64）：市场 ID

- `UserAddress`（string）：用户钱包地址（小写存储）

- `YesAmount`（string）：YES 侧累计数量

- `NoAmount`（string）：NO 侧累计数量

- `Claimed`（bool）：是否已 claim

- `Market`（*Market）：嵌套市场详情

- `UpdatedAt`（time.Time）：最后更新时间

---

### 8.4 User 结构体

#### 结构体说明

- 平台用户，以钱包地址为主键标识。

#### 规范与约定

- DID 绑定后为 did:pkh 格式。

#### 组合/依赖关系

- SIWE 登录 Upsert 创建/更新。

#### 结构体字段

- `ID`（int64）：用户主键

- `Address`（string）：以太坊地址

- `DID`（*string）：绑定的 DID，未绑定为 nil

---

## 9. internal/config（环境配置）

#### 模块说明

- 从环境变量加载全局 Config。

#### 规范与约定

- DATABASE_URL 必填；CHAIN_ID、INDEXER_START_BLOCK 必须可解析。

- BlockedCountries 逗号分隔转大写 map。

#### 组合/依赖关系

- 被所有 cmd 与 internal 包引用。



### 9.1 Config 结构体

#### 结构体说明

- 应用全局配置载体。

#### 规范与约定

- 敏感项：JWTSecret、OraclePrivateKey、AdminAPIKey、VCIssuerKey。

- 合约地址空字符串表示功能关闭。

#### 组合/依赖关系

- 通过 Load() 构造，注入 server/handler/indexer/oracle 等。

#### 结构体字段

- `HTTPPort`（string）：HTTP 监听端口，默认 8080

- `DatabaseURL`（string）：PostgreSQL 连接串，必填

- `RedisURL`（string）：Redis URL

- `EthRPCURL`（string）：以太坊 JSON-RPC

- `EthRPCFallback`（string）：备用 RPC

- `ChainID`（int64）：期望链 ID

- `JWTSecret`（string）：JWT HS256 密钥

- `SIWEDomain`（string）：SIWE domain

- `SIWEURI`（string）：SIWE uri

- `FactoryAddress`（string）：MarketFactory 地址

- `FactoryV3Address`（string）：MarketFactoryV3 地址

- `CollateralAddress`（string）：MockUSDC 地址

- `OracleAdapterAddress`（string）：Oracle Adapter 地址

- `OracleAdapterV2`（string）：Oracle Adapter V2 地址

- `DIDRegistryAddress`（string）：DID 注册表地址

- `OraclePrivateKey`（string）：发链上交易私钥

- `IndexerStartBlock`（uint64）：索引起始区块

- `IndexerPollSeconds`（int）：索引轮询秒数

- `MockMatchesPath`（string）：主 mock 比赛 JSON 路径

- `MockMatchesSecondary`（string）：次 mock JSON 路径

- `OracleCooldownMinutes`（int）：赛后创建 job 冷却分钟

- `OracleTimelockSeconds`（int）：链上 timelock 秒数，≤0 用 ResolveNow

- `OraclePollSeconds`（int）：Oracle Worker 轮询秒数

- `AdminAPIKey`（string）：管理端 API Key

- `VCIssuerKey`（string）：VC HMAC 密钥

- `AlertWebhookURL`（string）：告警 Webhook

- `BlockedCountries`（map[string]bool）：封禁国家码

- `GeoBlockEnabled`（bool）：是否启用地理围栏

- `RateLimitPerMinute`（int）：每 IP 每分钟限流

- `KYCWebhookSecret`（string）：KYC Webhook HMAC 密钥

- `ComplianceRequired`（bool）：合规开关

- `Environment`（string）：环境名 development 等

---

### 9.2 Load

- **函数用途**

  - 从环境变量构造并校验 Config。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `(*Config, error)`：成功返回配置指针

- **函数内校验**

  - DATABASE_URL 非空；CHAIN_ID、INDEXER_START_BLOCK 可解析。

- **函数实现效果**

  - 填充默认值 → parseBlockedCountries → 解析 chainID/startBlock

- **错误返回**

  - DATABASE_URL 缺失、CHAIN_ID/INDEXER_START_BLOCK 无效返回 error

---

### 9.3 parseBlockedCountries

- **函数用途**

  - 解析 BLOCKED_COUNTRIES 环境变量。

- **函数参数说明**

  - `raw`（string）：逗号分隔国家码

- **函数返回参数说明**

  - `map[string]bool`：大写国家码 → true

- **函数内校验**

  - 忽略空段。

- **函数实现效果**

  - Split → Trim → ToUpper → 写入 map

- **错误返回**

  - 无 error 返回

---

### 9.4 getEnv

- **函数用途**

  - 读环境变量，空则 fallback。

- **函数参数说明**

  - `key`（string）：变量名

  - `fallback`（string）：默认值

- **函数返回参数说明**

  - `string`：最终值

- **函数内校验**

  - 无。

- **函数实现效果**

  - os.Getenv 或 fallback

- **错误返回**

  - 无

---

### 9.5 getEnvInt

- **函数用途**

  - 读整型环境变量。

- **函数参数说明**

  - `key`（string）：变量名

  - `fallback`（int）：默认整型

- **函数返回参数说明**

  - `int`：解析值或 fallback

- **函数内校验**

  - Atoi 失败用 fallback。

- **函数实现效果**

  - strconv.Atoi

- **错误返回**

  - 无

---

## 10. internal/database（PostgreSQL）

#### 模块说明

- 连接池、Ping、迁移。

#### 规范与约定

- Ping 超时 3s；ErrNoChange 视为迁移成功。

#### 组合/依赖关系

- 被 cmd/* 与 server 使用。



### 10.1 NewPool

- **函数用途**

  - 创建 pgxpool。

- **函数参数说明**

  - `ctx`

  - `databaseURL`

- **函数返回参数说明**

  - `*pgxpool.Pool`

  - `error`

- **函数内校验**

  - 无

- **函数实现效果**

  - pgxpool.New

- **错误返回**

  - 连接失败 error

---

### 10.2 Ping

- **函数用途**

  - 检测 DB 可用。

- **函数参数说明**

  - `ctx`

  - `pool`

- **函数返回参数说明**

  - `error`

- **函数内校验**

  - 3s 超时

- **函数实现效果**

  - pool.Ping

- **错误返回**

  - 超时/失败 error

---

### 10.3 RunMigrations

- **函数用途**

  - golang-migrate Up。

- **函数参数说明**

  - `databaseURL`

  - `migrationsPath`

- **函数返回参数说明**

  - `error`

- **函数内校验**

  - 无

- **函数实现效果**

  - file:// 迁移 Up

- **错误返回**

  - 创建 migrator/Up 失败 error

---

## 11. internal/redis

#### 模块说明

- Redis 客户端封装。

#### 规范与约定

- Ping 2s 超时。

#### 组合/依赖关系

- server.Health、cmd/api 可选。



### 11.1 NewClient

- **函数用途**

  - ParseURL 创建 client。

- **函数参数说明**

  - `redisURL`

- **函数返回参数说明**

  - `*redis.Client`

  - `error`

- **函数内校验**

  - URL 合法

- **函数实现效果**

  - redis.ParseURL

- **错误返回**

  - 解析失败 error

---

### 11.2 Ping

- **函数用途**

  - Redis 连通性。

- **函数参数说明**

  - `ctx`

  - `client`

- **函数返回参数说明**

  - `error`

- **函数内校验**

  - 2s 超时

- **函数实现效果**

  - client.Ping

- **错误返回**

  - 失败 error

---

## 12. internal/auth（JWT / SIWE / 中间件）

#### 模块说明

- 身份认证与 HTTP 中间件。

#### 规范与约定

- JWT 仅 HS256；地址统一 lower。

- DID 绑定 MVP 不验链上签名。

#### 组合/依赖关系

- handler 路由分组使用 Middleware/AdminMiddleware。



### 12.1 Claims 结构体

#### 结构体说明

- JWT 载荷。

#### 规范与约定

- 嵌入 jwt.RegisteredClaims。

#### 组合/依赖关系

- ParseJWT 解析目标。

#### 结构体字段

- `Address`（string，json:`address`）：用户钱包地址小写

- `RegisteredClaims`（jwt.RegisteredClaims）：exp/iat 等标准声明

---

### 12.2 SIWEConfig 结构体

#### 结构体说明

- SIWE 域与 URI 配置。

#### 规范与约定

- 与 config.SIWEDomain/SIWEURI 对应。

#### 组合/依赖关系

- 传入 VerifySIWE。

#### 结构体字段

- `Domain`（string）：SIWE 消息 domain

- `URI`（string）：SIWE 消息 uri

---

### 12.3 IssueJWT

- **函数用途**

  - 签发 HS256 JWT。

- **函数参数说明**

  - `secret`（string）：密钥

  - `address`（string）：钱包地址

  - `ttl`（time.Duration）：有效期

- **函数返回参数说明**

  - `string`：JWT 字符串

  - `error`：签名失败

- **函数内校验**

  - 地址转 lower。

- **函数实现效果**

  - NewWithClaims → SignedString

- **错误返回**

  - 签名 error

---

### 12.4 ParseJWT

- **函数用途**

  - 解析校验 JWT。

- **函数参数说明**

  - `secret`

  - `tokenStr`

- **函数返回参数说明**

  - `*Claims`

  - `error`

- **函数内校验**

  - 算法必须 HS256；token.Valid。

- **函数实现效果**

  - ParseWithClaims

- **错误返回**

  - parse 失败、算法不符、invalid token

---

### 12.5 VerifySIWE

- **函数用途**

  - 验证 SIWE 消息与签名。

- **函数参数说明**

  - `cfg`（SIWEConfig）

  - `message`（string）

  - `signature`（string）

- **函数返回参数说明**

  - `string`：验证通过地址 lower

  - `error`

- **函数内校验**

  - domain/uri 匹配；未过期；签名有效。

- **函数实现效果**

  - siwe.ParseMessage → Verify

- **错误返回**

  - parse/domain/uri/expired/verify 错误

---

### 12.6 VerifyDIDBind

- **函数用途**

  - 校验 DID 格式。

- **函数参数说明**

  - `chainID`

  - `address`

  - `did`

  - `signatureHex`（未使用）

- **函数返回参数说明**

  - `error`

- **函数内校验**

  - did 必须等于 did:pkh:eip155:{chainID}:{address}。

- **函数实现效果**

  - 字符串比对

- **错误返回**

  - 格式不符 error

---

### 12.7 Middleware

- **函数用途**

  - JWT Bearer 中间件。

- **函数参数说明**

  - `secret`（string）

- **函数返回参数说明**

  - `func(http.Handler) http.Handler`

- **函数内校验**

  - Authorization Bearer 前缀；JWT 有效。

- **函数实现效果**

  - ParseJWT → context 写入 AddressKey

- **错误返回**

  - 无 Bearer 401 unauthorized；无效 401 invalid token

---

### 12.8 AddressFromContext

- **函数用途**

  - 从 context 取地址。

- **函数参数说明**

  - `ctx`（context.Context）

- **函数返回参数说明**

  - `string`：地址或空

- **函数内校验**

  - 类型断言 Claims。

- **函数实现效果**

  - ctx.Value(AddressKey)

- **错误返回**

  - 无 error

---

### 12.9 AdminMiddleware

- **函数用途**

  - 管理员 API Key 中间件。

- **函数参数说明**

  - `apiKey`（string）

- **函数返回参数说明**

  - `func(http.Handler) http.Handler`

- **函数内校验**

  - apiKey 空 503；X-Admin-Key 或 Bearer 匹配。

- **函数实现效果**

  - 比对 Header

- **错误返回**

  - 503 admin not configured；403 forbidden

---

## 13. internal/middleware

#### 模块说明

- 限流与地理围栏。

#### 规范与约定

- /health /ready 豁免限流。

#### 组合/依赖关系

- server.New 挂载。



### 13.1 ipLimiter 结构体

#### 结构体说明

- 按 IP 滑动窗口计数（包内私有）。

#### 规范与约定

- 每分钟重置 windowAt。

#### 组合/依赖关系

- RateLimit 内部使用。

#### 结构体字段

- `mu`（sync.Mutex）：并发锁

- `counts`（map[string]int）：IP → 计数

- `windowAt`（time.Time）：当前窗口起始

- `limit`（int）：每窗口上限

---

### 13.2 newIPLimiter

- **函数用途**

  - 创建 limiter。

- **函数参数说明**

  - `perMinute`（int）

- **函数返回参数说明**

  - `*ipLimiter`

- **函数内校验**

  - limit>0 由调用方保证

- **函数实现效果**

  - 初始化 map

- **错误返回**

  - 无

---

### 13.3 allow

- **函数用途**

  - IP 是否未超限。

- **函数参数说明**

  - `ip`（string）

- **函数返回参数说明**

  - `bool`

- **函数内校验**

  - 窗口过期重置 counts

- **函数实现效果**

  - 递增计数

- **错误返回**

  - 超限 false

---

### 13.4 RateLimit

- **函数用途**

  - 全局限流中间件。

- **函数参数说明**

  - `perMinute`（int）

- **函数返回参数说明**

  - 中间件函数

- **函数内校验**

  - health/ready 豁免

- **函数实现效果**

  - 429 rate_limited

- **错误返回**

  - 无

---

### 13.5 GeoBlock

- **函数用途**

  - 地理围栏中间件。

- **函数参数说明**

  - `cfg`

  - `logFn`

- **函数返回参数说明**

  - 中间件

- **函数内校验**

  - GeoBlockEnabled；exempt path

- **函数实现效果**

  - 403 region_restricted

- **错误返回**

  - 无

---

### 13.6 isExemptPath

- **函数用途**

  - 路径是否豁免。

- **函数参数说明**

  - `path`

- **函数返回参数说明**

  - `bool`

- **函数内校验**

  - 固定前缀列表

- **函数实现效果**

  - 匹配 exempt

- **错误返回**

  - 无

---

### 13.7 detectCountry

- **函数用途**

  - 从 Header 推断国家。

- **函数参数说明**

  - `r`

- **函数返回参数说明**

  - `string`

- **函数内校验**

  - CF-IPCountry 优先

- **函数实现效果**

  - 大写或 UNKNOWN

- **错误返回**

  - 无

---

## 14. internal/alert

#### 模块说明

- Webhook 告警。

#### 规范与约定

- webhook 空仅 log。

#### 组合/依赖关系

- oracle.Worker 使用。



### 14.1 Notifier 结构体

#### 结构体说明

- 告警通知器。

#### 规范与约定

- HTTP 5s 超时。

#### 组合/依赖关系

- New 构造；Send 发告警。

#### 结构体字段

- `webhookURL`（string）：目标 URL

- `client`（*http.Client）：HTTP 客户端

---

### 14.2 New

- **函数用途**

  - 构造 Notifier。

- **函数参数说明**

  - `webhookURL`

- **函数返回参数说明**

  - `*Notifier`

- **函数内校验**

  - 无

- **函数实现效果**

  - 5s timeout client

- **错误返回**

  - 无

---

### 14.3 Send

- **函数用途**

  - 发告警。

- **函数参数说明**

  - `event`

  - `message`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - log + 可选 POST JSON

- **错误返回**

  - POST 失败仅 log

---

## 15. internal/blockchain

#### 模块说明

- RPC 健康与 Oracle 链上交易。

#### 规范与约定

- Oracle ABI 从 pkg/contracts 加载。

#### 组合/依赖关系

- server、indexer、oracle、reconcile 使用。



### 15.1 Client 结构体

#### 结构体说明

- RPC 健康检查客户端。

#### 规范与约定

- atomic 存 rpcOK/chainID。

#### 组合/依赖关系

- Health.ready 读取 RPCOK/ChainID。

#### 结构体字段

- `url`（string）：RPC URL

- `expectedID`（int64）：期望链 ID，0 不校验

- `rpcOK`（atomic.Bool）：最近 Ping 是否成功

- `chainID`（atomic.Int64）：最近读到的链 ID

---

### 15.2 OracleClient 结构体

#### 结构体说明

- Oracle Adapter 写操作客户端。

#### 规范与约定

- 私钥本地签名发交易。

#### 组合/依赖关系

- RequestResolve/ConfirmResolve/ResolveNow/VoidMarket。

#### 结构体字段

- `client`（*ethclient.Client）：RPC

- `adapter`（common.Address）：Adapter 合约

- `abi`（abi.ABI）：合约 ABI

- `chainID`（*big.Int）：链 ID

- `auth`（*bind.TransactOpts）：交易签名选项

---

### 15.3 New

- **函数用途**

  - 构造 RPC Client。

- **函数参数说明**

  - `url`

  - `expectedChainID`

- **函数返回参数说明**

  - `*Client`

- **函数内校验**

  - 无

- **函数实现效果**

  - 保存 url/expectedID

- **错误返回**

  - 无

---

### 15.4 StartBackgroundPing

- **函数用途**

  - 后台 30s Ping。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - ctx 取消停止

- **函数实现效果**

  - pingOnce 循环

- **错误返回**

  - 无

---

### 15.5 pingOnce

- **函数用途**

  - 单次 RPC Ping。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 5s 超时；chainId 比对

- **函数实现效果**

  - 更新 rpcOK

- **错误返回**

  - Dial/ChainID 失败置 false

---

### 15.6 RPCOK

- **函数用途**

  - RPC 是否可用。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - `bool`

- **函数内校验**

  - 无

- **函数实现效果**

  - Load rpcOK

- **错误返回**

  - 无

---

### 15.7 ChainID

- **函数用途**

  - 最近链 ID。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - `int64`

- **函数内校验**

  - 无

- **函数实现效果**

  - Load chainID

- **错误返回**

  - 无

---

### 15.8 NewOracleClient

- **函数用途**

  - 构造 Oracle 客户端。

- **函数参数说明**

  - `ctx`

  - `rpcURL`

  - `adapterAddr`

  - `privateKeyHex`

  - `chainID`

- **函数返回参数说明**

  - `*OracleClient`

  - `error`

- **函数内校验**

  - 私钥 hex 可带 0x

- **函数实现效果**

  - Dial+ABI+ECDSA+bindOpts

- **错误返回**

  - Dial/ABI/私钥/bind 错误

---

### 15.9 bindOpts

- **函数用途**

  - TransactOpts。

- **函数参数说明**

  - `client`

  - `pk`

  - `chainID`

- **函数返回参数说明**

  - `*bind.TransactOpts`

  - `error`

- **函数内校验**

  - 无

- **函数实现效果**

  - NewKeyedTransactorWithChainID+GasPrice

- **错误返回**

  - 创建失败 error

---

### 15.10 loadAdapterABI

- **函数用途**

  - 加载 OracleAdapter.json。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - `abi.ABI`

  - `error`

- **函数内校验**

  - 多路径尝试

- **函数实现效果**

  - 读 JSON abi 字段

- **错误返回**

  - 文件/解析 error

---

### 15.11 RequestResolve

- **函数用途**

  - requestResolve 交易。

- **函数参数说明**

  - `ctx`

  - `market`

  - `outcome`

- **函数返回参数说明**

  - `string` txHash

  - `error`

- **函数内校验**

  - market hex

- **函数实现效果**

  - transact

- **错误返回**

  - transact 错误

---

### 15.12 ConfirmResolve

- **函数用途**

  - confirmResolve。

- **函数参数说明**

  - `ctx`

  - `market`

- **函数返回参数说明**

  - txHash

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - transact

- **错误返回**

  - transact 错误

---

### 15.13 ResolveNow

- **函数用途**

  - resolveNow。

- **函数参数说明**

  - `ctx`

  - `market`

  - `outcome`

- **函数返回参数说明**

  - txHash

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - transact

- **错误返回**

  - transact 错误

---

### 15.14 VoidMarket

- **函数用途**

  - voidMarket。

- **函数参数说明**

  - `ctx`

  - `market`

- **函数返回参数说明**

  - txHash

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - transact

- **错误返回**

  - transact 错误

---

### 15.15 transact

- **函数用途**

  - 通用写交易。

- **函数参数说明**

  - `ctx`

  - `method`

  - `args...`

- **函数返回参数说明**

  - txHash

  - error

- **函数内校验**

  - Pack 成功

- **函数实现效果**

  - Pack→EstimateGas(失败默认500000)→Sign→SendTransaction

- **错误返回**

  - Pack/Nonce/Sign/Send 错误

---

### 15.16 WaitMined

- **函数用途**

  - 等待交易上链。

- **函数参数说明**

  - `ctx`

  - `txHash`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 最多 30×2s

- **函数实现效果**

  - TransactionReceipt；status=0 revert

- **错误返回**

  - reverted/ctx.Err/timeout

---

### 15.17 ERC20Balance

- **函数用途**

  - balanceOf。

- **函数参数说明**

  - `ctx`

  - `rpcURL`

  - `tokenAddr`

  - `holder`

- **函数返回参数说明**

  - `*big.Int`

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - Dial+CallContract

- **错误返回**

  - Dial/Call 错误

---

## 16. internal/vc

#### 模块说明

- W3C VC HMAC 签发/校验。

#### 规范与约定

- 默认 TTL 365 天。

#### 组合/依赖关系

- handler credentials、kyc 使用。



### 16.1 Issuer 结构体

#### 结构体说明

- VC 签发器。

#### 规范与约定

- HMAC-SHA256 proof。

#### 组合/依赖关系

- NewIssuer 注入 key。

#### 结构体字段

- `key`（string）：HMAC 密钥

---

### 16.2 IssueRequest 结构体

#### 结构体说明

- 签发请求。

#### 规范与约定

- Claims nil 时用空 map 并设 id=SubjectDID。

#### 组合/依赖关系

- Issue 入参。

#### 结构体字段

- `SubjectDID`（string）：主体 DID

- `Type`（string）：凭证类型名

- `Claims`（map[string]interface{}）：主体声明

- `TTL`（time.Duration）：有效期，0 则 365 天

---

### 16.3 NewIssuer

- **函数用途**

  - 构造 Issuer。

- **函数参数说明**

  - `key`

- **函数返回参数说明**

  - `*Issuer`

- **函数内校验**

  - 无

- **函数实现效果**

  - 保存 key

- **错误返回**

  - 无

---

### 16.4 Issue

- **函数用途**

  - 生成 VC JSON。

- **函数参数说明**

  - `req`（IssueRequest）

- **函数返回参数说明**

  - `json.RawMessage`

  - error

- **函数内校验**

  - TTL 默认

- **函数实现效果**

  - Marshal VC+proof

- **错误返回**

  - Marshal error

---

### 16.5 sign

- **函数用途**

  - HMAC 签名。

- **函数参数说明**

  - `payload`（[]byte）

- **函数返回参数说明**

  - `string` base64

- **函数内校验**

  - 无

- **函数实现效果**

  - HMAC-SHA256

- **错误返回**

  - 无

---

### 16.6 Verify

- **函数用途**

  - 校验 VC。

- **函数参数说明**

  - `raw`

- **函数返回参数说明**

  - error

- **函数内校验**

  - proof 存在；签名匹配；未过期

- **函数实现效果**

  - 去 proof 重算 sign

- **错误返回**

  - missing proof/invalid signature/expired

---

### 16.7 SubjectRegion

- **函数用途**

  - 取 region。

- **函数参数说明**

  - `raw`

- **函数返回参数说明**

  - string

  - error

- **函数内校验**

  - credentialSubject 存在

- **函数实现效果**

  - Upper region

- **错误返回**

  - unmarshal/no subject

---

## 17. internal/repository（数据访问层）

#### 模块说明

- PostgreSQL CRUD，无 ORM。

#### 规范与约定

- 地址字段统一 lower；numeric 用 string。

#### 组合/依赖关系

- handler、indexer、oracle、wcprovider 注入各 Repo。



### 17.1 MatchRepo 结构体

#### 结构体说明

- 比赛表访问。

#### 规范与约定

- pool 注入。

#### 组合/依赖关系

- NewMatchRepo 构造。

#### 结构体字段

- `pool`（*pgxpool.Pool）：连接池

---

### 17.2 NewMatchRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*MatchRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.3 List

- **函数用途**

  - 分页列表。

- **函数参数说明**

  - `ctx`

  - `status`

  - `limit`

  - `offset`

- **函数返回参数说明**

  - []models.Match

  - error

- **函数内校验**

  - limit≤0→20

- **函数实现效果**

  - SELECT ORDER kickoff

- **错误返回**

  - Query/Scan error

---

### 17.4 GetByID

- **函数用途**

  - 按 ID 查。

- **函数参数说明**

  - `ctx`

  - `id`

- **函数返回参数说明**

  - *models.Match

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - QueryRow

- **错误返回**

  - 无行 error

---

### 17.5 Upsert

- **函数用途**

  - 按 external_id upsert。

- **函数参数说明**

  - `ctx`

  - `m`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - INSERT ON CONFLICT

- **错误返回**

  - Exec error

---

### 17.6 SetStatus

- **函数用途**

  - 更新 status。

- **函数参数说明**

  - `ctx`

  - `id`

  - `status`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.7 GetByExternalID

- **函数用途**

  - 按 external_id。

- **函数参数说明**

  - `ctx`

  - `externalID`

- **函数返回参数说明**

  - *models.Match

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - QueryRow

- **错误返回**

  - 无行 error

---

### 17.8 ParseKickoff

- **函数用途**

  - RFC3339 解析。

- **函数参数说明**

  - `s`

- **函数返回参数说明**

  - time.Time

  - error

- **函数内校验**

  - RFC3339

- **函数实现效果**

  - time.Parse

- **错误返回**

  - 格式 error

---

### 17.9 MarketRepo 结构体

#### 结构体说明

- markets 表。

#### 规范与约定

- JOIN matches。

#### 组合/依赖关系

- handler/indexer 使用。

#### 结构体字段

- `pool`（*pgxpool.Pool）

---

### 17.10 AdminMarketUpdate 结构体

#### 结构体说明

- 管理员更新合规字段。

#### 规范与约定

- ResolutionRule 空默认 HOME_WIN。

#### 组合/依赖关系

- RegisterAdmin 入参。

#### 结构体字段

- `MatchID`（int64）：比赛 ID

- `RequiresVC`（bool）：是否需 VC

- `RestrictedRegion`（string）：限制地区，空写 NULL

- `ResolutionRule`（string）：结算规则

---

### 17.11 NewMarketRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*MarketRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.12 List

- **函数用途**

  - 分页+JOIN。

- **函数参数说明**

  - `ctx`

  - `status`

  - `limit`

  - `offset`

- **函数返回参数说明**

  - []models.Market

  - error

- **函数内校验**

  - limit≤0→20

- **函数实现效果**

  - SELECT LEFT JOIN

- **错误返回**

  - Query error

---

### 17.13 GetByID

- **函数用途**

  - 详情+比赛。

- **函数参数说明**

  - `ctx`

  - `id`

- **函数返回参数说明**

  - *models.Market

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - scanMarketRow

- **错误返回**

  - Scan error

---

### 17.14 GetByAddress

- **函数用途**

  - 按合约地址。

- **函数参数说明**

  - `ctx`

  - `addr`

- **函数返回参数说明**

  - *models.Market

  - error

- **函数内校验**

  - lower 比较

- **函数实现效果**

  - QueryRow

- **错误返回**

  - 无行 error

---

### 17.15 InsertFromChain

- **函数用途**

  - Indexer 写入。

- **函数参数说明**

  - `ctx`

  - `mk`

- **函数返回参数说明**

  - int64 id

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - INSERT ON CONFLICT market_address

- **错误返回**

  - Exec error

---

### 17.16 UpdateResolved

- **函数用途**

  - 标记 RESOLVED。

- **函数参数说明**

  - `ctx`

  - `marketAddress`

  - `outcome`

  - `yesPool`

  - `noPool`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.17 ListOpenByMatchID

- **函数用途**

  - 某场 OPEN 市场。

- **函数参数说明**

  - `ctx`

  - `matchID`

- **函数返回参数说明**

  - []models.Market

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - List OPEN 后过滤

- **错误返回**

  - List error

---

### 17.18 RegisterAdmin

- **函数用途**

  - 更新合规字段。

- **函数参数说明**

  - `ctx`

  - `req`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE markets BY match_id

- **错误返回**

  - Exec error

---

### 17.19 SetVoid

- **函数用途**

  - status=VOID。

- **函数参数说明**

  - `ctx`

  - `id`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.20 UpdatePools

- **函数用途**

  - 更新 yes/no pool。

- **函数参数说明**

  - `ctx`

  - `marketAddress`

  - `yesPool`

  - `noPool`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.21 scanMarketWithMatch

- **函数用途**

  - 扫描联表行。

- **函数参数说明**

  - `rows`

- **函数返回参数说明**

  - models.Market

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - Scan

- **错误返回**

  - Scan error

---

### 17.22 scanMarketRow

- **函数用途**

  - 扫描单行。

- **函数参数说明**

  - `row`

- **函数返回参数说明**

  - *models.Market

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - Scan+嵌套 Match

- **错误返回**

  - Scan error

---

### 17.23 derefStr

- **函数用途**

  - *string 解引用。

- **函数参数说明**

  - `s`

- **函数返回参数说明**

  - string

- **函数内校验**

  - nil→""

- **函数实现效果**

  - 解引用

- **错误返回**

  - 无

---

### 17.24 derefTime

- **函数用途**

  - *time.Time 解引用。

- **函数参数说明**

  - `t`

- **函数返回参数说明**

  - time.Time

- **函数内校验**

  - nil→零值

- **函数实现效果**

  - 解引用

- **错误返回**

  - 无

---

### 17.25 PositionRepo 结构体

#### 结构体说明

- 持仓与 trades。

#### 规范与约定

- AddTrade UPSERT 累加。

#### 组合/依赖关系

- indexer/handler。

#### 结构体字段

- `pool`

---

### 17.26 NewPositionRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*PositionRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.27 AddTrade

- **函数用途**

  - 累加持仓。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `userAddress`

  - `outcome`

  - `amount`

- **函数返回参数说明**

  - error

- **函数内校验**

  - outcome 0=yes

- **函数实现效果**

  - UPSERT positions

- **错误返回**

  - Exec error

---

### 17.28 SetClaimed

- **函数用途**

  - claimed=true。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `userAddress`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.29 ListByUser

- **函数用途**

  - 用户持仓列表。

- **函数参数说明**

  - `ctx`

  - `address`

- **函数返回参数说明**

  - []models.Position

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - JOIN markets

- **错误返回**

  - Query error

---

### 17.30 InsertTrade

- **函数用途**

  - 写 trades 幂等。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `txHash`

  - `logIndex`

  - `blockNumber`

  - `userAddress`

  - `outcome`

  - `amount`

- **函数返回参数说明**

  - error

- **函数内校验**

  - tx_hash+log_index 冲突忽略

- **函数实现效果**

  - INSERT

- **错误返回**

  - Exec error

---

### 17.31 UserRepo 结构体

#### 结构体说明

- users 表。

#### 规范与约定

- 地址 lower。

#### 组合/依赖关系

- auth handler。

#### 结构体字段

- `pool`

---

### 17.32 NewUserRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*UserRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.33 UpsertByAddress

- **函数用途**

  - 插入或 touch。

- **函数参数说明**

  - `ctx`

  - `address`

- **函数返回参数说明**

  - *models.User

  - error

- **函数内校验**

  - lower address

- **函数实现效果**

  - INSERT ON CONFLICT

- **错误返回**

  - Exec/Scan error

---

### 17.34 BindDID

- **函数用途**

  - 绑定 DID。

- **函数参数说明**

  - `ctx`

  - `address`

  - `did`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE did

- **错误返回**

  - Exec error

---

### 17.35 GetByAddress

- **函数用途**

  - 按地址查。

- **函数参数说明**

  - `ctx`

  - `address`

- **函数返回参数说明**

  - *models.User

  - error

- **函数内校验**

  - lower

- **函数实现效果**

  - QueryRow

- **错误返回**

  - 无行 error

---

### 17.36 Credential 结构体

#### 结构体说明

- DB 凭证记录。

#### 规范与约定

- 与 models 分离，含 Revoked/CreatedAt。

#### 组合/依赖关系

- CredentialRepo 使用。

#### 结构体字段

- `ID`（int64）

- `UserAddress`（string）

- `CredentialType`（string）

- `VCJSON`（json.RawMessage）

- `ExpiresAt`（time.Time）

- `Revoked`（bool）

- `CreatedAt`（time.Time）

---

### 17.37 CredentialRepo 结构体

#### 结构体说明

- credentials 表。

#### 规范与约定

- HasValidType 查未 revoke 未过期。

#### 组合/依赖关系

- handler VC。

#### 结构体字段

- `pool`

---

### 17.38 NewCredentialRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*CredentialRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.39 Insert

- **函数用途**

  - 插入凭证。

- **函数参数说明**

  - `ctx`

  - `c`

- **函数返回参数说明**

  - int64 id

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - INSERT RETURNING id

- **错误返回**

  - Exec error

---

### 17.40 ListByUser

- **函数用途**

  - 用户有效凭证。

- **函数参数说明**

  - `ctx`

  - `address`

- **函数返回参数说明**

  - []Credential

  - error

- **函数内校验**

  - 未 revoke 未过期

- **函数实现效果**

  - SELECT

- **错误返回**

  - Query error

---

### 17.41 HasValidType

- **函数用途**

  - 是否存在有效类型。

- **函数参数说明**

  - `ctx`

  - `address`

  - `credType`

- **函数返回参数说明**

  - bool

  - error

- **函数内校验**

  - EXISTS

- **函数实现效果**

  - QueryRow

- **错误返回**

  - Scan error

---

### 17.42 ComplianceRepo 结构体

#### 结构体说明

- 合规日志。

#### 规范与约定

- geo/kyc 表。

#### 组合/依赖关系

- middleware GeoBlock 回调。

#### 结构体字段

- `pool`

---

### 17.43 NewComplianceRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*ComplianceRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.44 LogGeo

- **函数用途**

  - 写 geo_access_log。

- **函数参数说明**

  - `ctx`

  - `ip`

  - `country`

  - `path`

  - `allowed`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - INSERT

- **错误返回**

  - Exec error

---

### 17.45 LogKYC

- **函数用途**

  - 写 kyc_events。

- **函数参数说明**

  - `ctx`

  - `externalID`

  - `userAddr`

  - `status`

  - `raw`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - INSERT

- **错误返回**

  - Exec error

---

### 17.46 IndexerStateRepo 结构体

#### 结构体说明

- indexer_state 单行。

#### 规范与约定

- id=1。

#### 组合/依赖关系

- indexer 使用。

#### 结构体字段

- `pool`

---

### 17.47 NewIndexerStateRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*IndexerStateRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.48 GetLastBlock

- **函数用途**

  - 读 last_block。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - uint64

  - error

- **函数内校验**

  - id=1

- **函数实现效果**

  - QueryRow

- **错误返回**

  - Scan error

---

### 17.49 SetLastBlock

- **函数用途**

  - 更新 last_block。

- **函数参数说明**

  - `ctx`

  - `block`

  - `factory`

- **函数返回参数说明**

  - error

- **函数内校验**

  - factory 空保留原值

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.50 OracleJob 结构体

#### 结构体说明

- 预言机作业。

#### 规范与约定

- status: pending/submitted/manual_review/confirmed/failed。

#### 组合/依赖关系

- oracle worker + admin API。

#### 结构体字段

- `ID`（int64）

- `MatchID`（*int64）

- `MarketID`（int64）

- `Status`（string）

- `PrimaryHome/Away`（*int）：主源比分

- `SecondaryHome/Away`（*int）：次源比分

- `ProposedOutcome`（*int）

- `TxHash`（*string）

- `ErrorMessage`（*string）

- `ExecuteAfter`（time.Time）：冷却后可执行时间

- `CreatedAt/UpdatedAt`（time.Time）

- `MarketAddress/Question`（string）：JOIN 字段

---

### 17.51 OracleJobRepo 结构体

#### 结构体说明

- oracle_jobs 表。

#### 规范与约定

- UpdateStatus COALESCE 保留 nil 字段。

#### 组合/依赖关系

- oracle/handler。

#### 结构体字段

- `pool`

---

### 17.52 NewOracleJobRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*OracleJobRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.53 Create

- **函数用途**

  - 创建 pending job。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `matchID`

  - `executeAfter`

- **函数返回参数说明**

  - int64 id

  - error

- **函数内校验**

  - status=pending

- **函数实现效果**

  - INSERT RETURNING

- **错误返回**

  - Scan error

---

### 17.54 HasActiveForMarket

- **函数用途**

  - 是否有活跃 job。

- **函数参数说明**

  - `ctx`

  - `marketID`

- **函数返回参数说明**

  - bool

  - error

- **函数内校验**

  - pending/submitted/manual_review

- **函数实现效果**

  - EXISTS

- **错误返回**

  - Scan error

---

### 17.55 ListDue

- **函数用途**

  - 到期 pending。

- **函数参数说明**

  - `ctx`

  - `now`

- **函数返回参数说明**

  - []OracleJob

  - error

- **函数内校验**

  - execute_after<=now

- **函数实现效果**

  - JOIN markets

- **错误返回**

  - Query error

---

### 17.56 ListAll

- **函数用途**

  - 分页列表。

- **函数参数说明**

  - `ctx`

  - `status`

  - `limit`

- **函数返回参数说明**

  - []OracleJob

  - error

- **函数内校验**

  - limit≤0→50

- **函数实现效果**

  - 可选 status 过滤

- **错误返回**

  - Query error

---

### 17.57 UpdateStatus

- **函数用途**

  - 更新 job。

- **函数参数说明**

  - `ctx`

  - `id`

  - `status`

  - `fields`

- **函数返回参数说明**

  - error

- **函数内校验**

  - COALESCE 字段

- **函数实现效果**

  - UPDATE

- **错误返回**

  - Exec error

---

### 17.58 scanOracleJobRow

- **函数用途**

  - 扫描 job 行。

- **函数参数说明**

  - `rows`

- **函数返回参数说明**

  - OracleJob

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - Scan

- **错误返回**

  - Scan error

---

### 17.59 PlatformStats 结构体

#### 结构体说明

- 平台统计 DTO。

#### 规范与约定

- fees 估算 volume*0.003。

#### 组合/依赖关系

- Stats.Platform 返回。

#### 结构体字段

- `TradeCount`（int64）

- `TradeVolume`（string）

- `FeesCollected`（string）

- `ActiveUsers`（int64）

- `OpenMarkets`（int64）

- `TVLApprox`（string）

---

### 17.60 StatsRepo 结构体

#### 结构体说明

- 统计 SQL。

#### 规范与约定

- 无

#### 组合/依赖关系

- handler stats。

#### 结构体字段

- `pool`

---

### 17.61 NewStatsRepo

- **函数用途**

  - 构造。

- **函数参数说明**

  - `pool`

- **函数返回参数说明**

  - `*StatsRepo`

- **函数内校验**

  - 无

- **函数实现效果**

  - 返回 struct

- **错误返回**

  - 无

---

### 17.62 Platform

- **函数用途**

  - 聚合统计。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - *PlatformStats

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - 多表 COUNT/SUM

- **错误返回**

  - Query error

---

### 17.63 UpdateMarketPool

- **函数用途**

  - 更新 reserve/price。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `yes`

  - `no`

  - `priceBps`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - UPDATE markets

- **错误返回**

  - Exec error

---

## 18. internal/handler（HTTP 处理器）

#### 模块说明

- REST/SSE/Admin 路由实现。

#### 规范与约定

- 统一 writeJSON/writeError；parseID 解析 chi {id}。

#### 组合/依赖关系

- API 聚合各 Repo + VCIssuer + OracleChain。



### 18.1 API 结构体

#### 结构体说明

- HTTP API 依赖容器。

#### 规范与约定

- RegisterRoutes 注册全部路由。

#### 组合/依赖关系

- server.New 注入各字段。

#### 结构体字段

- `Cfg`（*config.Config）

- `Matches/Markets/Users/Positions/OracleJobs/Credentials/Stats/Compliance`（各 *Repo）

- `VCIssuer`（*vc.Issuer）

- `OracleChain`（*blockchain.OracleClient，可 nil）

---

### 18.2 Health 结构体

#### 结构体说明

- 健康检查。

#### 规范与约定

- ready DB 失败 503。

#### 组合/依赖关系

- server 注册。

#### 结构体字段

- `DB`

- `Redis`

- `Chain`

---

### 18.3 siweRequest 结构体

#### 结构体说明

- SIWE 请求体。

#### 规范与约定

- json tag。

#### 组合/依赖关系

- siweAuth。

#### 结构体字段

- `Message`

- `Signature`

---

### 18.4 bindDIDRequest 结构体

#### 结构体说明

- DID 绑定体。

#### 规范与约定

- 无

#### 组合/依赖关系

- bindDID。

#### 结构体字段

- `DID`

- `Signature`

---

### 18.5 issueVCRequest 结构体

#### 结构体说明

- 签发 VC 体。

#### 规范与约定

- address+type 必填。

#### 组合/依赖关系

- issueCredential。

#### 结构体字段

- `Address`

- `CredentialType`

- `Claims`

- `TTLHours`

---

### 18.6 verifyVCRequest 结构体

#### 结构体说明

- 校验 VC 体。

#### 规范与约定

- 无

#### 组合/依赖关系

- verifyVC。

#### 结构体字段

- `VCJSON`

- `CredentialType`

- `Region`

---

### 18.7 voidMarketRequest 结构体

#### 结构体说明

- 作废市场体。

#### 规范与约定

- reason 未持久化。

#### 组合/依赖关系

- voidMarket。

#### 结构体字段

- `Reason`

---

### 18.8 registerMarketRequest 结构体

#### 结构体说明

- 登记市场体。

#### 规范与约定

- MarketAddress/Question 未直接用。

#### 组合/依赖关系

- registerMarket。

#### 结构体字段

- `MatchID`

- `MarketAddress`

- `Question`

- `RequiresVC`

- `RestrictedRegion`

- `ResolutionRule`

---

### 18.9 RegisterRoutes

- **函数用途**

  - 注册 chi 路由。

- **函数参数说明**

  - `r`（chi.Router）

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JWT/Admin 分组

- **函数实现效果**

  - 挂载全部 handler

- **错误返回**

  - 无

---

### 18.10 writeJSON

- **函数用途**

  - JSON 响应。

- **函数参数说明**

  - `w`

  - `status`

  - `v`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - Content-Type+Encode

- **错误返回**

  - Encode 忽略 error

---

### 18.11 writeError

- **函数用途**

  - 错误 JSON。

- **函数参数说明**

  - `w`

  - `status`

  - `msg`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - {error:msg}

- **错误返回**

  - 无

---

### 18.12 queryInt

- **函数用途**

  - query 整型。

- **函数参数说明**

  - `r`

  - `key`

  - `def`

- **函数返回参数说明**

  - int

- **函数内校验**

  - Atoi 失败用 def

- **函数实现效果**

  - ParseInt

- **错误返回**

  - 无

---

### 18.13 parseID

- **函数用途**

  - URL id。

- **函数参数说明**

  - `r`

- **函数返回参数说明**

  - int64

  - error

- **函数内校验**

  - chi URLParam id

- **函数实现效果**

  - ParseInt

- **错误返回**

  - 非数字 error

---

### 18.14 health

- **函数用途**

  - 存活探针。

- **函数参数说明**

  - `w`

  - `_`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - 200 {status:ok}

- **错误返回**

  - 无

---

### 18.15 ready

- **函数用途**

  - 就绪探针。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - DB ping 失败 503

- **函数实现效果**

  - 返回 db/redis/rpc 状态

- **错误返回**

  - 503 not ready

---

### 18.16 siweAuth

- **函数用途**

  - POST /auth/siwe。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JSON body；VerifySIWE

- **函数实现效果**

  - Upsert+JWT 24h

- **错误返回**

  - 400/401/500

---

### 18.17 bindDID

- **函数用途**

  - POST /users/bind-did。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JWT；VerifyDIDBind

- **函数实现效果**

  - BindDID→user

- **错误返回**

  - 400/500

---

### 18.18 myPositions

- **函数用途**

  - GET /me/positions。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JWT

- **函数实现效果**

  - ListByUser

- **错误返回**

  - 500

---

### 18.19 listMatches

- **函数用途**

  - GET /matches。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - status/limit/offset

- **函数实现效果**

  - List

- **错误返回**

  - 500

---

### 18.20 getMatch

- **函数用途**

  - GET /matches/{id}。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - parseID

- **函数实现效果**

  - GetByID+markets

- **错误返回**

  - 400/404/500

---

### 18.21 listMarkets

- **函数用途**

  - GET /markets。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - query

- **函数实现效果**

  - List+collateral+chain_id

- **错误返回**

  - 500

---

### 18.22 getMarket

- **函数用途**

  - GET /markets/{id}。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - RequiresVC→HasValidType VerifiedFan

- **函数实现效果**

  - market+access gate

- **错误返回**

  - 400/404/500

---

### 18.23 platformStats

- **函数用途**

  - GET /stats/platform。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - Stats.Platform

- **错误返回**

  - 500

---

### 18.24 marketPool

- **函数用途**

  - GET /markets/{id}/pool。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - parseID

- **函数实现效果**

  - 返回 reserve/price/fee

- **错误返回**

  - 400/404/500

---

### 18.25 marketOrderbook

- **函数用途**

  - GET /markets/{id}/orderbook。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - parseBps 默认 5000

- **函数实现效果**

  - 合成 CPMM bids

- **错误返回**

  - 400/404/500

---

### 18.26 parseBps

- **函数用途**

  - 解析 bps。

- **函数参数说明**

  - `s`

- **函数返回参数说明**

  - int

- **函数内校验**

  - 空或0→5000

- **函数实现效果**

  - Atoi

- **错误返回**

  - 无

---

### 18.27 coalesceReserve

- **函数用途**

  - 选储备字符串。

- **函数参数说明**

  - `primary`

  - `fallback`

- **函数返回参数说明**

  - string

- **函数内校验**

  - primary 空或0用 fallback

- **函数实现效果**

  - 比较

- **错误返回**

  - 无

---

### 18.28 issueCredential

- **函数用途**

  - POST /credentials/issue admin。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - address+type 必填

- **函数实现效果**

  - Issue+Insert

- **错误返回**

  - 400/500

---

### 18.29 myCredentials

- **函数用途**

  - GET /users/me/credentials。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JWT

- **函数实现效果**

  - ListByUser

- **错误返回**

  - 500

---

### 18.30 verifyVC

- **函数用途**

  - POST /auth/verify-vc。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - Verify；region 可选

- **函数实现效果**

  - {valid:true}

- **错误返回**

  - 401/403

---

### 18.31 complianceRestricted

- **函数用途**

  - GET /compliance/restricted。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - detectCountry

- **函数实现效果**

  - restricted 状态

- **错误返回**

  - 无

---

### 18.32 kycWebhook

- **函数用途**

  - POST /kyc/webhook。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - body≤1MB；HMAC 可选；external_id+status

- **函数实现效果**

  - LogKYC；approved 签发 VC（未入库）

- **错误返回**

  - 400/401/500

---

### 18.33 listOracleJobs

- **函数用途**

  - GET /admin/oracle-jobs。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - Admin

- **函数实现效果**

  - ListAll limit100

- **错误返回**

  - 500

---

### 18.34 voidMarket

- **函数用途**

  - POST /admin/markets/{id}/void。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - parseID；GetByID

- **函数实现效果**

  - 链上 Void+SetVoid+match CANCELLED

- **错误返回**

  - 400/404/500

---

### 18.35 registerMarket

- **函数用途**

  - POST /admin/markets。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - JSON body

- **函数实现效果**

  - RegisterAdmin

- **错误返回**

  - 400/500

---

### 18.36 retryOracleJob

- **函数用途**

  - POST /admin/oracle-jobs/{id}/retry。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - parseID

- **函数实现效果**

  - status pending

- **错误返回**

  - 400

---

### 18.37 streamScores

- **函数用途**

  - GET /events/scores SSE。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - Flusher

- **函数实现效果**

  - 每5s推送 matches

- **错误返回**

  - 500 no flusher

---

### 18.38 prometheusMetrics

- **函数用途**

  - GET /metrics。

- **函数参数说明**

  - `w`

  - `r`

- **函数返回参数说明**

  - 无

- **函数内校验**

  - 无

- **函数实现效果**

  - oracle_jobs 计数 Prometheus 文本

- **错误返回**

  - 500

---

## 19. internal/server

#### 模块说明

- HTTP 服务器组装。

#### 规范与约定

- WriteTimeout=0 支持 SSE。

#### 组合/依赖关系

- cmd/api 调用 New。



### 19.1 Server 结构体

#### 结构体说明

- 封装 http.Server。

#### 规范与约定

- 无

#### 组合/依赖关系

- ListenAndServe/Shutdown。

#### 结构体字段

- `httpServer`（*http.Server）

---

### 19.2 Dependencies 结构体

#### 结构体说明

- New 注入依赖。

#### 规范与约定

- 无

#### 组合/依赖关系

- cmd/api 传入。

#### 结构体字段

- `Port`

- `Cfg`

- `DB`

- `Redis`

- `Chain`

- `OracleChain`

---

### 19.3 New

- **函数用途**

  - 组装 chi+中间件+路由。

- **函数参数说明**

  - `deps`（Dependencies）

- **函数返回参数说明**

  - `*Server`

- **函数内校验**

  - 无

- **函数实现效果**

  - RateLimit→GeoBlock→CORS→Health→API

- **错误返回**

  - 无

---

### 19.4 ListenAndServe

- **函数用途**

  - 阻塞监听。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - http.ListenAndServe

- **错误返回**

  - Listen 错误

---

### 19.5 Shutdown

- **函数用途**

  - 优雅关闭。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - http.Shutdown

- **错误返回**

  - Shutdown error

---

### 19.6 Addr

- **函数用途**

  - 监听地址。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - string

- **函数内校验**

  - 无

- **函数实现效果**

  - httpServer.Addr

- **错误返回**

  - 无

---

### 19.7 String

- **函数用途**

  - 日志 URL。

- **函数参数说明**

  - 无

- **函数返回参数说明**

  - string

- **函数内校验**

  - 无

- **函数实现效果**

  - http://localhost{Addr}

- **错误返回**

  - 无

---

## 20. internal/indexer

#### 模块说明

- 链上事件→DB。

#### 规范与约定

- Factory 空则 skip；每批最多 2000 块。

#### 组合/依赖关系

- cmd/api goroutine 或 cmd/indexer。



### 20.1 Indexer 结构体

#### 结构体说明

- 链索引器。

#### 规范与约定

- FilterLogs 扫 MarketCreated/Bought/Resolved/Claimed。

#### 组合/依赖关系

- 注入 matches/markets/positions/state repos。

#### 结构体字段

- `cfg`（*config.Config）

- `client`（*ethclient.Client）

- `factoryABI/marketABI`（abi.ABI）

- `matches/markets/positions/state`（各 *Repo）

---

### 20.2 New

- **函数用途**

  - 构造 Indexer。

- **函数参数说明**

  - `ctx`

  - `cfg`

  - repos...

- **函数返回参数说明**

  - *Indexer

  - error

- **函数内校验**

  - Dial+loadABI

- **函数实现效果**

  - Dial MarketFactory+PredictionMarket ABI

- **错误返回**

  - Dial/ABI error

---

### 20.3 loadABI

- **函数用途**

  - 读 contract JSON。

- **函数参数说明**

  - `name`

- **函数返回参数说明**

  - abi.ABI

  - error

- **函数内校验**

  - 多路径

- **函数实现效果**

  - Unmarshal abi

- **错误返回**

  - 文件 error

---

### 20.4 Run

- **函数用途**

  - 主循环 poll。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - FactoryAddress 空 return nil

- **函数实现效果**

  - ticker IndexerPollSeconds

- **错误返回**

  - ctx.Err

---

### 20.5 poll

- **函数用途**

  - 单轮扫描。

- **函数参数说明**

  - `ctx`

  - `factoryAddr`

- **函数返回参数说明**

  - error

- **函数内校验**

  - last>=head return

- **函数实现效果**

  - scanFactory+scanMarkets+SetLastBlock

- **错误返回**

  - state/head/scan error

---

### 20.6 scanFactory

- **函数用途**

  - MarketCreated。

- **函数参数说明**

  - `ctx`

  - `factory`

  - `from`

  - `to`

- **函数返回参数说明**

  - error

- **函数内校验**

  - FilterLogs

- **函数实现效果**

  - handleMarketCreated

- **错误返回**

  - Filter error

---

### 20.7 handleMarketCreated

- **函数用途**

  - 写入 market。

- **函数参数说明**

  - `ctx`

  - `lg`

- **函数返回参数说明**

  - error

- **函数内校验**

  - Topics>=4

- **函数实现效果**

  - matchRef→external；InsertFromChain

- **错误返回**

  - Unpack/Insert error

---

### 20.8 matchRefToExternal

- **函数用途**

  - 哈希→external_id。

- **函数参数说明**

  - `matchRefHex`

- **函数返回参数说明**

  - string

- **函数内校验**

  - 已知映射

- **函数实现效果**

  - keccak 比对

- **错误返回**

  - 无

---

### 20.9 scanMarkets

- **函数用途**

  - 各市场 Bought/Resolved/Claimed。

- **函数参数说明**

  - `ctx`

  - `from`

  - `to`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 最多500市场

- **函数实现效果**

  - FilterLogs per market

- **错误返回**

  - List/Filter error

---

### 20.10 handleBought

- **函数用途**

  - 交易+持仓。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `lg`

- **函数返回参数说明**

  - error

- **函数内校验**

  - Unpack

- **函数实现效果**

  - InsertTrade+AddTrade

- **错误返回**

  - Unpack/DB error

---

### 20.11 handleResolved

- **函数用途**

  - UpdateResolved。

- **函数参数说明**

  - `ctx`

  - `marketAddress`

  - `lg`

- **函数返回参数说明**

  - error

- **函数内校验**

  - Unpack

- **函数实现效果**

  - UpdateResolved

- **错误返回**

  - Unpack/DB error

---

### 20.12 handleClaimed

- **函数用途**

  - SetClaimed。

- **函数参数说明**

  - `ctx`

  - `marketID`

  - `lg`

- **函数返回参数说明**

  - error

- **函数内校验**

  - Topics[1] user

- **函数实现效果**

  - SetClaimed

- **错误返回**

  - DB error

---

## 21. internal/oracle

#### 模块说明

- 赛后自动结算 Worker。

#### 规范与约定

- 双源不一致→manual_review+告警。

#### 组合/依赖关系

- cmd/oracle。



### 21.1 Worker 结构体

#### 结构体说明

- Oracle 后台 Worker。

#### 规范与约定

- Timelock≤0 用 ResolveNow，否则 request+confirm。

#### 组合/依赖关系

- 注入 jobs/matches/markets/dual/chain/alerts。

#### 结构体字段

- `cfg`

- `jobs`

- `matches`

- `markets`

- `dual`

- `chain`（可 nil）

- `alerts`

---

### 21.2 NewWorker

- **函数用途**

  - 构造 Worker。

- **函数参数说明**

  - cfg

  - repos

  - dual

  - chain

  - alerts

- **函数返回参数说明**

  - *Worker

- **函数内校验**

  - 无

- **函数实现效果**

  - 赋值字段

- **错误返回**

  - 无

---

### 21.3 Run

- **函数用途**

  - 定时 tick。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - OraclePollSeconds ticker

- **函数实现效果**

  - tick 循环

- **错误返回**

  - ctx.Err

---

### 21.4 tick

- **函数用途**

  - enqueue+processDue。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - 顺序调用

- **错误返回**

  - enqueue/process error

---

### 21.5 enqueueFinished

- **函数用途**

  - FINISHED 比赛入队。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - 无 active job

- **函数实现效果**

  - Create job+cooldown；match ORACLE_PENDING

- **错误返回**

  - List/Create error

---

### 21.6 processDue

- **函数用途**

  - 处理到期 job。

- **函数参数说明**

  - `ctx`

- **函数返回参数说明**

  - error

- **函数内校验**

  - chain nil skip

- **函数实现效果**

  - ListDue→processJob

- **错误返回**

  - ListDue error

---

### 21.7 processJob

- **函数用途**

  - 单 job 结算。

- **函数参数说明**

  - `ctx`

  - `job`（OracleJob）

- **函数返回参数说明**

  - error

- **函数内校验**

  - MatchID nil return；CompareScores

- **函数实现效果**

  - 不一致→manual_review；一致→OutcomeFromRule→链上 resolve→UpdateResolved→match RESOLVED

- **错误返回**

  - Compare/GetByID/链上 error；不一致 return nil

---

### 21.8 merge

- **函数用途**

  - 合并 map。

- **函数参数说明**

  - `base`

  - `extra`

- **函数返回参数说明**

  - map

- **函数内校验**

  - 无

- **函数实现效果**

  - extra 写入 base

- **错误返回**

  - 无

---

## 22. internal/wcprovider

#### 模块说明

- 世界杯 Mock 数据源。

#### 规范与约定

- JSON RFC3339 kickoff。

#### 组合/依赖关系

- sync/oracle/seed。



### 22.1 MockProvider 结构体

#### 结构体说明

- 单文件 mock。

#### 规范与约定

- 无

#### 组合/依赖关系

- Sync 导入 DB。

#### 结构体字段

- `path`（string）：JSON 路径

---

### 22.2 mockMatch 结构体

#### 结构体说明

- JSON 行结构（未导出）。

#### 规范与约定

- json tag 映射。

#### 组合/依赖关系

- MockProvider.Sync 使用。

#### 结构体字段

- ExternalID

- HomeTeam

- AwayTeam

- KickoffAt

- Status

- HomeScore

- AwayScore

---

### 22.3 ScoreSnapshot 结构体

#### 结构体说明

- 比分快照。

#### 规范与约定

- CompareScores 返回。

#### 组合/依赖关系

- DualProvider。

#### 结构体字段

- ExternalID

- HomeScore

- AwayScore

- Status

---

### 22.4 DualProvider 结构体

#### 结构体说明

- 双源 Provider。

#### 规范与约定

- primary/secondary 路径。

#### 组合/依赖关系

- CompareScores/SyncPrimary。

#### 结构体字段

- `primaryPath`

- `secondaryPath`

---

### 22.5 NewMock

- **函数用途**

  - 构造 MockProvider。

- **函数参数说明**

  - `path`

- **函数返回参数说明**

  - *MockProvider

- **函数内校验**

  - 无

- **函数实现效果**

  - 保存 path

- **错误返回**

  - 无

---

### 22.6 Sync

- **函数用途**

  - 读 JSON Upsert。

- **函数参数说明**

  - `ctx`

  - `repo`

- **函数返回参数说明**

  - int count

  - error

- **函数内校验**

  - kickoff RFC3339

- **函数实现效果**

  - 逐条 Upsert

- **错误返回**

  - 读文件/解析/Upsert error

---

### 22.7 NewDual

- **函数用途**

  - 构造 DualProvider。

- **函数参数说明**

  - `primary`

  - `secondary`

- **函数返回参数说明**

  - *DualProvider

- **函数内校验**

  - 无

- **函数实现效果**

  - 保存路径

- **错误返回**

  - 无

---

### 22.8 load

- **函数用途**

  - 加载有比分的条目。

- **函数参数说明**

  - `path`

- **函数返回参数说明**

  - map[string]ScoreSnapshot

  - error

- **函数内校验**

  - home/away score 均存在

- **函数实现效果**

  - 读 JSON

- **错误返回**

  - 文件/JSON error

---

### 22.9 CompareScores

- **函数用途**

  - 双源比对。

- **函数参数说明**

  - `externalID`

- **函数返回参数说明**

  - primary

  - secondary

  - match bool

  - err

- **函数内校验**

  - 两侧均有数据且比分一致 match=true

- **函数实现效果**

  - load 双文件

- **错误返回**

  - load error

---

### 22.10 SyncPrimary

- **函数用途**

  - 仅同步主源。

- **函数参数说明**

  - `ctx`

  - `repo`

- **函数返回参数说明**

  - int

  - error

- **函数内校验**

  - 无

- **函数实现效果**

  - MockProvider.Sync primary

- **错误返回**

  - Sync error

---

### 22.11 OutcomeFromRule

- **函数用途**

  - 规则→outcome 0/1。

- **函数参数说明**

  - `rule`

  - `home`

  - `away`

- **函数返回参数说明**

  - int

- **函数内校验**

  - OVER_25 或 HOME_WIN

- **函数实现效果**

  - 总进球或主胜

- **错误返回**

  - 无

---

### 22.12 FinishMatch

- **函数用途**

  - 设 FINISHED+比分 Upsert。

- **函数参数说明**

  - `ctx`

  - `repo`

  - `externalID`

  - `home`

  - `away`

- **函数返回参数说明**

  - error

- **函数内校验**

  - GetByExternalID

- **函数实现效果**

  - Upsert

- **错误返回**

  - Get/Upsert error

---

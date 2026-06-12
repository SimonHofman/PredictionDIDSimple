# Backend 函数说明文档

本文档按模块整理 backend 目录下各 Go 源文件中的函数，包含函数用途、参数说明与返回值说明。

---

## 1. cmd/api

### 1.1 main

- **函数用途**
  - HTTP API 服务主入口，加载配置、执行数据库迁移、初始化 PostgreSQL/Redis/区块链客户端，可选内嵌 Indexer，组装 HTTP Server 并监听请求，支持优雅关闭。

- **函数参数说明**
  - 无参数（程序入口函数）。

- **返回参数说明**
  - 无返回值；初始化失败或服务器异常退出时调用 `log.Fatalf` 终止进程。

---

## 2. cmd/indexer

### 2.1 main

- **函数用途**
  - 独立链上索引器进程入口，连接数据库并启动 Indexer，持续扫描链上事件写入 DB。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；错误时 `log.Fatal` 退出。

---

## 3. cmd/oracle

### 3.1 main

- **函数用途**
  - 预言机 Worker 进程入口，初始化 Oracle 链上客户端与 Worker，循环处理赛后自动结算任务。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非取消类错误时 `log.Fatal` 退出。

---

## 4. cmd/sync

### 4.1 main

- **函数用途**
  - 一次性同步 Worker，从 Mock 主数据源 JSON 文件同步世界杯赛程到数据库。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；同步失败时 `log.Fatal` 退出，成功时打印同步条数。

---

## 5. cmd/reconcile

### 5.1 main

- **函数用途**
  - 链上/DB 对账工具入口，加载配置、迁移数据库后执行对账逻辑。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；缺少 `MOCK_USDC_ADDRESS` 或对账失败时终止进程。

### 5.2 run

- **函数用途**
  - 遍历 OPEN/RESOLVED 状态市场，比较 DB 中池子总量与链上 ERC20 抵押余额，将结果写入 `reconciliation_runs` 表。

- **函数参数说明**
  - `ctx`（`context.Context`）：请求上下文，用于数据库查询与链上调用。
  - `pool`（`*pgxpool.Pool`）：PostgreSQL 连接池。
  - `cfg`（`*config.Config`）：应用配置，含 RPC URL、抵押代币地址等。

- **返回参数说明**
  - `error`：查询、扫描或迭代过程中出错时返回错误；全部市场处理完毕返回 `rows.Err()`。

---

## 6. cmd/migrate

### 6.1 main

- **函数用途**
  - 独立数据库迁移入口，加载配置并对 PostgreSQL 执行 `migrations/` 下的 Up 迁移。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；迁移失败时 `log.Fatalf` 退出。

---

## 7. cmd/seed

### 7.1 main

- **函数用途**
  - 种子数据入口，执行迁移后从 Mock JSON 文件导入比赛数据到数据库。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；失败时 `log.Fatal` 退出，成功时打印导入条数。

---

## 8. internal/config

### 8.1 Load

- **函数用途**
  - 从环境变量加载应用配置，解析链 ID、索引起始区块、封禁国家列表等，校验 `DATABASE_URL` 必填。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `*Config`：解析成功的配置结构体指针。
  - `error`：`DATABASE_URL` 缺失、`CHAIN_ID` 或 `INDEXER_START_BLOCK` 解析失败时返回错误。

### 8.2 parseBlockedCountries

- **函数用途**
  - 将逗号分隔的国家代码字符串解析为大写国家码 → `true` 的 map，用于地理围栏。

- **函数参数说明**
  - `raw`（`string`）：逗号分隔的国家代码，如 `"US,KP,IR,SY"`。

- **返回参数说明**
  - `map[string]bool`：非空国家码为 key、值为 `true` 的 map。

### 8.3 getEnv

- **函数用途**
  - 读取环境变量，未设置或为空时返回默认值。

- **函数参数说明**
  - `key`（`string`）：环境变量名。
  - `fallback`（`string`）：默认值。

- **返回参数说明**
  - `string`：环境变量值或 fallback。

### 8.4 getEnvInt

- **函数用途**
  - 读取整型环境变量，未设置、为空或解析失败时返回默认值。

- **函数参数说明**
  - `key`（`string`）：环境变量名。
  - `fallback`（`int`）：默认整数值。

- **返回参数说明**
  - `int`：解析后的整数或 fallback。

---

## 9. internal/database

### 9.1 NewPool

- **函数用途**
  - 根据数据库 URL 创建 PostgreSQL 连接池（pgxpool）。

- **函数参数说明**
  - `ctx`（`context.Context`）：连接上下文。
  - `databaseURL`（`string`）：PostgreSQL 连接字符串。

- **返回参数说明**
  - `*pgxpool.Pool`：连接池实例。
  - `error`：连接失败时返回包装后的错误。

### 9.2 Ping

- **函数用途**
  - 检测数据库连接是否可用，超时 3 秒。

- **函数参数说明**
  - `ctx`（`context.Context`）：父上下文。
  - `pool`（`*pgxpool.Pool`）：待检测的连接池。

- **返回参数说明**
  - `error`：Ping 失败或超时时返回错误，成功返回 `nil`。

### 9.3 RunMigrations

- **函数用途**
  - 使用 golang-migrate 对指定路径的 SQL 迁移文件执行 Up 操作。

- **函数参数说明**
  - `databaseURL`（`string`）：PostgreSQL 连接字符串。
  - `migrationsPath`（`string`）：迁移文件目录路径（不含 `file://` 前缀）。

- **返回参数说明**
  - `error`：创建 migrator 或执行 Up 失败时返回错误；无新迁移（`ErrNoChange`）视为成功。

---

## 10. internal/redis

### 10.1 NewClient

- **函数用途**
  - 解析 Redis URL 并创建 go-redis 客户端。

- **函数参数说明**
  - `redisURL`（`string`）：Redis 连接 URL，如 `redis://localhost:6379/0`。

- **返回参数说明**
  - `*redis.Client`：Redis 客户端实例。
  - `error`：URL 解析失败时返回错误。

### 10.2 Ping

- **函数用途**
  - 检测 Redis 连接是否可用，超时 2 秒。

- **函数参数说明**
  - `ctx`（`context.Context`）：父上下文。
  - `client`（`*redis.Client`）：Redis 客户端。

- **返回参数说明**
  - `error`：Ping 失败时返回错误，成功返回 `nil`。

---

## 11. internal/auth

### 11.1 IssueJWT

- **函数用途**
  - 为指定钱包地址签发 HS256 JWT，Claims 含小写 `address` 与过期/签发时间。

- **函数参数说明**
  - `secret`（`string`）：JWT 签名密钥。
  - `address`（`string`）：用户以太坊地址（会转为小写写入 Claims）。
  - `ttl`（`time.Duration`）：令牌有效期。

- **返回参数说明**
  - `string`：签名后的 JWT 字符串。
  - `error`：签名失败时返回错误。

### 11.2 ParseJWT

- **函数用途**
  - 解析并校验 JWT，提取 Claims；仅接受 HS256 算法。

- **函数参数说明**
  - `secret`（`string`）：JWT 验证密钥。
  - `tokenStr`（`string`）：待解析的 JWT 字符串。

- **返回参数说明**
  - `*Claims`：解析成功且有效的 Claims（含 `Address` 等字段）。
  - `error`：解析失败、算法不匹配或 token 无效时返回错误。

### 11.3 VerifySIWE

- **函数用途**
  - 验证 Sign-In with Ethereum（SIWE）消息与签名，校验 domain、URI、过期时间，返回验证通过的钱包地址。

- **函数参数说明**
  - `cfg`（`SIWEConfig`）：SIWE 配置，含 `Domain` 与 `URI`。
  - `message`（`string`）：客户端提交的 SIWE 消息文本。
  - `signature`（`string`）：对应签名的十六进制字符串。

- **返回参数说明**
  - `string`：验证通过的小写以太坊地址。
  - `error`：解析、domain/URI 不匹配、过期或验签失败时返回错误。

### 11.4 VerifyDIDBind

- **函数用途**
  - 校验用户绑定的 DID 格式是否为 `did:pkh:eip155:{chainID}:{address}`；MVP 阶段不校验链上签名。

- **函数参数说明**
  - `chainID`（`int64`）：链 ID。
  - `address`（`string`）：用户钱包地址。
  - `did`（`string`）：待绑定的 DID 字符串。
  - `signatureHex`（`string`）：客户端签名（当前实现中未使用）。

- **返回参数说明**
  - `error`：DID 与期望格式不一致时返回错误，否则返回 `nil`。

### 11.5 Middleware

- **函数用途**
  - 返回 HTTP 中间件：从 `Authorization: Bearer <token>` 解析 JWT，将地址写入请求 Context。

- **函数参数说明**
  - `secret`（`string`）：JWT 验证密钥。

- **返回参数说明**
  - `func(http.Handler) http.Handler`：Chi/标准库兼容的中间件函数；无 Bearer 或 token 无效时返回 401 JSON。

### 11.6 AddressFromContext

- **函数用途**
  - 从请求 Context 中读取经 JWT 中间件注入的用户地址。

- **函数参数说明**
  - `ctx`（`context.Context`）：HTTP 请求上下文。

- **返回参数说明**
  - `string`：用户地址；未注入或类型断言失败时返回空字符串。

### 11.7 AdminMiddleware

- **函数用途**
  - 返回管理员鉴权中间件，校验 `X-Admin-Key` 或 `Authorization: Bearer` 是否与配置的 Admin API Key 一致。

- **函数参数说明**
  - `apiKey`（`string`）：管理员 API 密钥；为空时所有请求返回 503。

- **返回参数说明**
  - `func(http.Handler) http.Handler`：中间件函数；密钥不匹配返回 403。

---

## 12. internal/middleware

### 12.1 newIPLimiter

- **函数用途**
  - 创建基于 IP 的滑动窗口限流器（每分钟重置计数）。

- **函数参数说明**
  - `perMinute`（`int`）：每个 IP 每分钟允许的最大请求数。

- **返回参数说明**
  - `*ipLimiter`：限流器实例。

### 12.2 allow（ipLimiter 方法）

- **函数用途**
  - 判断指定 IP 在当前分钟窗口内是否未超过限流阈值，并递增计数。

- **函数参数说明**
  - `ip`（`string`）：客户端 IP 地址。

- **返回参数说明**
  - `bool`：允许请求为 `true`，超限为 `false`。

### 12.3 RateLimit

- **函数用途**
  - 返回全局限流 HTTP 中间件；`/health` 与 `/ready` 不受限。

- **函数参数说明**
  - `perMinute`（`int`）：每 IP 每分钟请求上限。

- **返回参数说明**
  - `func(http.Handler) http.Handler`：中间件；超限时返回 429 JSON `rate_limited`。

### 12.4 GeoBlock

- **函数用途**
  - 返回地理围栏中间件，根据配置封禁国家列表拦截请求，并可回调记录访问日志。

- **函数参数说明**
  - `cfg`（`*config.Config`）：配置，含 `GeoBlockEnabled` 与 `BlockedCountries`。
  - `logFn`（`func(ip, country, path string, allowed bool)`）：可选日志回调；为 `nil` 时不记录。

- **返回参数说明**
  - `func(http.Handler) http.Handler`：中间件；封禁国家返回 403 JSON `region_restricted`。

### 12.5 isExemptPath

- **函数用途**
  - 判断请求路径是否豁免地理围栏（健康检查、指标、合规查询、SSE 事件流等）。

- **函数参数说明**
  - `path`（`string`）：HTTP 请求路径。

- **返回参数说明**
  - `bool`：豁免为 `true`，否则为 `false`。

### 12.6 detectCountry

- **函数用途**
  - 从 HTTP 请求头推断国家代码，优先 `CF-IPCountry`，其次 `X-Country-Code`，均无则 `UNKNOWN`。

- **函数参数说明**
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - `string`：大写国家代码或 `"UNKNOWN"`。

---

## 13. internal/alert

### 13.1 New

- **函数用途**
  - 创建告警通知器，用于向 Webhook 发送 JSON 告警（可选）。

- **函数参数说明**
  - `webhookURL`（`string`）：告警 Webhook URL；为空则仅写本地日志。

- **返回参数说明**
  - `*Notifier`：通知器实例，内置 5 秒超时的 HTTP 客户端。

### 13.2 Send（Notifier 方法）

- **函数用途**
  - 记录告警日志并向 Webhook POST `{"event","message"}` JSON；Webhook 为空或 POST 失败时仅打日志。

- **函数参数说明**
  - `event`（`string`）：告警事件类型标识。
  - `message`（`string`）：告警详情文本。

- **返回参数说明**
  - 无返回值。

---

## 14. internal/blockchain

### 14.1 New（Client）

- **函数用途**
  - 创建区块链 RPC 健康检查客户端，保存 RPC URL 与期望链 ID。

- **函数参数说明**
  - `url`（`string`）：以太坊 JSON-RPC URL。
  - `expectedChainID`（`int64`）：期望的链 ID，用于校验；为 0 时不校验。

- **返回参数说明**
  - `*Client`：RPC 客户端实例。

### 14.2 StartBackgroundPing（Client 方法）

- **函数用途**
  - 在后台 goroutine 中立即执行一次 RPC Ping，之后每 30 秒重复，直到 Context 取消。

- **函数参数说明**
  - `ctx`（`context.Context`）：取消信号上下文。

- **返回参数说明**
  - 无返回值。

### 14.3 pingOnce（Client 方法）

- **函数用途**
  - 单次 RPC 连通性检测：Dial、读取 ChainID、与期望值比对，更新 `rpcOK` 与 `chainID` 原子变量。

- **函数参数说明**
  - `ctx`（`context.Context`）：父上下文（内部创建 5 秒超时子上下文）。

- **返回参数说明**
  - 无返回值；失败时写 WARN 日志并置 `rpcOK` 为 false。

### 14.4 RPCOK（Client 方法）

- **函数用途**
  - 返回最近一次 Ping 是否成功。

- **函数参数说明**
  - 无参数（接收者为 `*Client`）。

- **返回参数说明**
  - `bool`：RPC 可用为 `true`。

### 14.5 ChainID（Client 方法）

- **函数用途**
  - 返回最近一次 Ping 读到的链 ID。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `int64`：链 ID；尚未 Ping 成功时可能为 0。

### 14.6 NewOracleClient

- **函数用途**
  - 创建 Oracle 链上交易客户端：连接 RPC、加载 OracleAdapter ABI、解析私钥并构造 TransactOpts。

- **函数参数说明**
  - `ctx`（`context.Context`）：Dial 上下文。
  - `rpcURL`（`string`）：以太坊 RPC URL。
  - `adapterAddr`（`string`）：Oracle Adapter 合约地址（十六进制）。
  - `privateKeyHex`（`string`）：发交易私钥（可带 `0x` 前缀）。
  - `chainID`（`int64`）：链 ID。

- **返回参数说明**
  - `*OracleClient`：Oracle 客户端实例。
  - `error`：Dial、ABI 加载、私钥解析或 TransactOpts 创建失败时返回错误。

### 14.7 bindOpts

- **函数用途**
  - 根据私钥与链 ID 创建 `bind.TransactOpts`，并尝试设置建议 Gas Price。

- **函数参数说明**
  - `client`（`*ethclient.Client`）：以太坊客户端，用于查询 Gas Price。
  - `pk`（`*ecdsa.PrivateKey`）：签名私钥。
  - `chainID`（`int64`）：链 ID。

- **返回参数说明**
  - `*bind.TransactOpts`：可用于合约写操作的选项。
  - `error`：创建 Transactor 失败时返回错误。

### 14.8 loadAdapterABI

- **函数用途**
  - 从 `pkg/contracts/OracleAdapter.json`（或 `backend/pkg/...`）加载 Oracle Adapter 合约 ABI。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `abi.ABI`：解析后的 ABI。
  - `error`：文件读取或 JSON 解析失败时返回错误。

### 14.9 RequestResolve（OracleClient 方法）

- **函数用途**
  - 调用 Oracle Adapter 的 `requestResolve(market, outcome)`，发起带时间锁的结算请求。

- **函数参数说明**
  - `ctx`（`context.Context`）：交易上下文。
  - `market`（`string`）：市场合约地址（十六进制）。
  - `outcome`（`uint8`）：提议的获胜结果（0=YES/主胜等，1=NO/客胜等，依规则而定）。

- **返回参数说明**
  - `string`：已发送交易的哈希（十六进制）。
  - `error`：打包、签名或发送交易失败时返回错误。

### 14.10 ConfirmResolve（OracleClient 方法）

- **函数用途**
  - 调用 `confirmResolve(market)`，在时间锁过后确认结算。

- **函数参数说明**
  - `ctx`（`context.Context`）：交易上下文。
  - `market`（`string`）：市场合约地址。

- **返回参数说明**
  - `string`：交易哈希。
  - `error`：交易失败时返回错误。

### 14.11 ResolveNow（OracleClient 方法）

- **函数用途**
  - 调用 `resolveNow(market, outcome)`，无时间锁立即结算市场。

- **函数参数说明**
  - `ctx`（`context.Context`）：交易上下文。
  - `market`（`string`）：市场合约地址。
  - `outcome`（`uint8`）：获胜结果。

- **返回参数说明**
  - `string`：交易哈希。
  - `error`：交易失败时返回错误。

### 14.12 VoidMarket（OracleClient 方法）

- **函数用途**
  - 调用 `voidMarket(market)`，作废市场。

- **函数参数说明**
  - `ctx`（`context.Context`）：交易上下文。
  - `market`（`string`）：市场合约地址。

- **返回参数说明**
  - `string`：交易哈希。
  - `error`：交易失败时返回错误。

### 14.13 transact（OracleClient 方法）

- **函数用途**
  - Oracle 写操作的通用实现：ABI Pack、估算 Gas、构造并签名交易、广播到链上。

- **函数参数说明**
  - `ctx`（`context.Context`）：交易上下文。
  - `method`（`string`）：合约方法名。
  - `args`（`...interface{}`）：方法参数列表。

- **返回参数说明**
  - `string`：已发送交易哈希。
  - `error`：Pack、Nonce、签名或 SendTransaction 失败时返回错误。

### 14.14 WaitMined（OracleClient 方法）

- **函数用途**
  - 轮询等待交易上链确认，最多约 60 秒；回执 status 为 0 视为 revert。

- **函数参数说明**
  - `ctx`（`context.Context`）：可取消的等待上下文。
  - `txHash`（`string`）：待等待的交易哈希。

- **返回参数说明**
  - `error`：交易 revert、Context 取消或超时返回错误；成功确认返回 `nil`。

### 14.15 ERC20Balance

- **函数用途**
  - 通过 RPC 调用 ERC20 `balanceOf(holder)`，查询指定地址在代币合约中的余额。

- **函数参数说明**
  - `ctx`（`context.Context`）：Call 上下文。
  - `rpcURL`（`string`）：以太坊 RPC URL。
  - `tokenAddr`（`string`）：ERC20 代币合约地址。
  - `holder`（`string`）：持仓地址。

- **返回参数说明**
  - `*big.Int`：代币余额（最小单位）。
  - `error`：Dial 或 CallContract 失败时返回错误。

---

## 15. internal/vc

### 15.1 NewIssuer

- **函数用途**
  - 创建可验证凭证（VC）签发器，使用 HMAC-SHA256 对凭证 payload 签名。

- **函数参数说明**
  - `key`（`string`）：HMAC 签名密钥。

- **返回参数说明**
  - `*Issuer`：签发器实例。

### 15.2 Issue（Issuer 方法）

- **函数用途**
  - 根据请求生成 W3C Verifiable Credential JSON，含 issuer、有效期、credentialSubject 与 HMAC proof。

- **函数参数说明**
  - `req`（`IssueRequest`）：签发请求。
    - `SubjectDID`（`string`）：主体 DID。
    - `Type`（`string`）：凭证类型名（写入 `type` 数组）。
    - `Claims`（`map[string]interface{}`）：主体声明字段；为 `nil` 时使用空 map，并自动设置 `id` 为 SubjectDID。
    - `TTL`（`time.Duration`）：有效期；为 0 时默认 365 天。

- **返回参数说明**
  - `json.RawMessage`：完整 VC JSON（含 proof）。
  - `error`：JSON 序列化失败时返回错误。

### 15.3 sign（Issuer 方法）

- **函数用途**
  - 对 payload 字节计算 HMAC-SHA256 并 Base64 编码，作为 proofValue。

- **函数参数说明**
  - `payload`（`[]byte`）：待签名的 JSON 字节（不含 proof 字段）。

- **返回参数说明**
  - `string`：Base64 编码的 HMAC 摘要。

### 15.4 Verify（Issuer 方法）

- **函数用途**
  - 校验 VC 的 HMAC 签名与过期时间。

- **函数参数说明**
  - `raw`（`json.RawMessage`）：完整 VC JSON。

- **返回参数说明**
  - `error`：JSON 无效、缺少 proof、签名不匹配或已过期时返回错误；成功返回 `nil`。

### 15.5 SubjectRegion

- **函数用途**
  - 从 VC JSON 的 `credentialSubject.region` 字段提取并转为大写地区码。

- **函数参数说明**
  - `raw`（`json.RawMessage`）：VC JSON。

- **返回参数说明**
  - `string`：大写 region 字符串；缺失时可能为空字符串。
  - `error`：JSON 解析失败或缺少 credentialSubject 时返回错误。

---

## 16. internal/repository

### 16.1 NewMatchRepo

- **函数用途**
  - 创建比赛数据仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*MatchRepo`：MatchRepo 实例。

### 16.2 List（MatchRepo 方法）

- **函数用途**
  - 分页查询比赛列表，可按 status 过滤，按开球时间升序。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `status`（`string`）：状态过滤；空字符串表示不过滤。
  - `limit`（`int`）：每页条数；≤0 时默认 20。
  - `offset`（`int`）：偏移量。

- **返回参数说明**
  - `[]models.Match`：比赛列表。
  - `error`：查询或 Scan 失败时返回错误。

### 16.3 GetByID（MatchRepo 方法）

- **函数用途**
  - 按主键 ID 查询单场比赛。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `id`（`int64`）：比赛 ID。

- **返回参数说明**
  - `*models.Match`：比赛记录。
  - `error`：未找到或数据库错误时返回错误。

### 16.4 Upsert（MatchRepo 方法）

- **函数用途**
  - 按 `external_id` 插入或更新比赛记录。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `m`（`models.Match`）：比赛实体，含 ExternalID、球队、开球时间、状态、比分等。

- **返回参数说明**
  - `error`：执行失败时返回错误。

### 16.5 SetStatus（MatchRepo 方法）

- **函数用途**
  - 更新指定比赛的状态字段。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `id`（`int64`）：比赛 ID。
  - `status`（`string`）：新状态，如 `FINISHED`、`ORACLE_PENDING`、`RESOLVED`。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.6 GetByExternalID（MatchRepo 方法）

- **函数用途**
  - 按外部 ID（如 `wc-2026-semi-001`）查询比赛。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `externalID`（`string`）：外部唯一标识。

- **返回参数说明**
  - `*models.Match`：比赛记录。
  - `error`：未找到或查询失败时返回错误。

### 16.7 ParseKickoff

- **函数用途**
  - 将 RFC3339 格式字符串解析为 `time.Time`（包级工具函数）。

- **函数参数说明**
  - `s`（`string`）：RFC3339 时间字符串。

- **返回参数说明**
  - `time.Time`：解析后的时间。
  - `error`：格式无效时返回错误。

### 16.8 NewMarketRepo

- **函数用途**
  - 创建市场数据仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*MarketRepo`：MarketRepo 实例。

### 16.9 List（MarketRepo 方法）

- **函数用途**
  - 分页查询市场列表（LEFT JOIN 比赛），可按 status 过滤。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `status`（`string`）：市场状态过滤；空表示全部。
  - `limit`（`int`）：每页条数；≤0 默认 20。
  - `offset`（`int`）：偏移量。

- **返回参数说明**
  - `[]models.Market`：市场列表（可含嵌套 Match）。
  - `error`：查询失败时返回错误。

### 16.10 GetByID（MarketRepo 方法）

- **函数用途**
  - 按市场主键 ID 查询详情（含关联比赛）。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `id`（`int64`）：市场 ID。

- **返回参数说明**
  - `*models.Market`：市场实体。
  - `error`：未找到或 Scan 失败时返回错误。

### 16.11 GetByAddress（MarketRepo 方法）

- **函数用途**
  - 按链上市场合约地址（大小写不敏感）查询市场。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `addr`（`string`）：市场合约地址。

- **返回参数说明**
  - `*models.Market`：市场实体。
  - `error`：未找到或查询失败时返回错误。

### 16.12 InsertFromChain（MarketRepo 方法）

- **函数用途**
  - 从链上索引写入市场记录；`market_address` 冲突时更新 `match_id`。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `mk`（`models.Market`）：市场实体，含 MatchID、ChainID、FactoryAddress、MarketAddress、OnChainMarketID、MatchRef、Question、EndTime、Status 等。

- **返回参数说明**
  - `int64`：插入或更新后的市场 ID。
  - `error`：执行失败时返回错误。

### 16.13 UpdateResolved（MarketRepo 方法）

- **函数用途**
  - 将市场标记为 RESOLVED，写入 winning_outcome 与 yes/no 池子。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketAddress`（`string`）：市场合约地址。
  - `outcome`（`int`）：获胜结果索引。
  - `yesPool`（`string`）：YES 池 numeric 字符串。
  - `noPool`（`string`）：NO 池 numeric 字符串。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.14 ListOpenByMatchID（MarketRepo 方法）

- **函数用途**
  - 查询指定比赛下所有 OPEN 状态的市场（内部先 List OPEN 再内存过滤 match_id）。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `matchID`（`int64`）：比赛 ID。

- **返回参数说明**
  - `[]models.Market`：该比赛下的开放市场列表。
  - `error`：查询失败时返回错误。

### 16.15 RegisterAdmin（MarketRepo 方法）

- **函数用途**
  - 管理员更新某比赛关联市场的合规与结算规则字段（requires_vc、restricted_region、resolution_rule）。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `req`（`AdminMarketUpdate`）：更新请求。
    - `MatchID`（`int64`）：比赛 ID。
    - `RequiresVC`（`bool`）：是否需要 VC 才能交易。
    - `RestrictedRegion`（`string`）：限制地区；空字符串写入 NULL。
    - `ResolutionRule`（`string`）：结算规则；空时默认 `HOME_WIN`。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.16 SetVoid（MarketRepo 方法）

- **函数用途**
  - 将指定市场状态设为 VOID。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `id`（`int64`）：市场 ID。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.17 UpdatePools（MarketRepo 方法）

- **函数用途**
  - 按市场地址更新 yes_pool 与 no_pool。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketAddress`（`string`）：市场合约地址。
  - `yesPool`（`string`）：YES 池 numeric 字符串。
  - `noPool`（`string`）：NO 池 numeric 字符串。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.18 scanMarketWithMatch

- **函数用途**
  - 从 `rows.Next()` 扫描一行市场+比赛联表数据为 `models.Market`。

- **函数参数说明**
  - `rows`：实现 `Next()` 与 `Scan(...)` 的行迭代器（如 pgx Rows）。

- **返回参数说明**
  - `models.Market`：扫描后的市场结构体。
  - `error`：Scan 失败时返回错误。

### 16.19 scanMarketRow

- **函数用途**
  - 从单行 QueryRow 结果扫描市场及可选嵌套 Match 信息。

- **函数参数说明**
  - `row`：实现 `Scan(...)` 的单行结果（如 pgx Row）。

- **返回参数说明**
  - `*models.Market`：市场指针，Match 非空时填充 `m.Match`。
  - `error`：Scan 失败时返回错误。

### 16.20 derefStr

- **函数用途**
  - 安全解引用 `*string`，nil 时返回空字符串。

- **函数参数说明**
  - `s`（`*string`）：可空字符串指针。

- **返回参数说明**
  - `string`：解引用值或 `""`。

### 16.21 derefTime

- **函数用途**
  - 安全解引用 `*time.Time`，nil 时返回零值时间。

- **函数参数说明**
  - `t`（`*time.Time`）：可空时间指针。

- **返回参数说明**
  - `time.Time`：解引用值或零值。

### 16.22 NewPositionRepo

- **函数用途**
  - 创建用户持仓仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*PositionRepo`：PositionRepo 实例。

### 16.23 AddTrade（PositionRepo 方法）

- **函数用途**
  - 累加用户在某个市场的 YES/NO 持仓（按 outcome 写入对应 side）。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `userAddress`（`string`）：用户地址（存储为小写）。
  - `outcome`（`int`）：结果侧，0=YES，非 0=NO。
  - `amount`（`string`）：交易数量 numeric 字符串。

- **返回参数说明**
  - `error`：Upsert 失败时返回错误。

### 16.24 SetClaimed（PositionRepo 方法）

- **函数用途**
  - 将用户在某市场的持仓标记为已领取（claimed=true）。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `userAddress`（`string`）：用户地址。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.25 ListByUser（PositionRepo 方法）

- **函数用途**
  - 查询某用户全部持仓，JOIN 市场信息，按更新时间降序。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `address`（`string`）：用户钱包地址。

- **返回参数说明**
  - `[]models.Position`：持仓列表，每项含嵌套 Market。
  - `error`：查询失败时返回错误。

### 16.26 InsertTrade（PositionRepo 方法）

- **函数用途**
  - 插入链上交易明细到 `trades` 表，`(tx_hash, log_index)` 冲突时忽略。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `txHash`（`string`）：交易哈希。
  - `logIndex`（`int`）：日志索引。
  - `blockNumber`（`int64`）：区块号。
  - `userAddress`（`string`）：用户地址。
  - `outcome`（`int`）：买卖方向/结果侧。
  - `amount`（`string`）：数量 numeric 字符串。

- **返回参数说明**
  - `error`：插入失败时返回错误。

### 16.27 NewUserRepo

- **函数用途**
  - 创建用户仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*UserRepo`：UserRepo 实例。

### 16.28 UpsertByAddress（UserRepo 方法）

- **函数用途**
  - 按地址插入用户或更新 `updated_at`，地址统一小写。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `address`（`string`）：以太坊地址。

- **返回参数说明**
  - `*models.User`：用户记录（id、address、did）。
  - `error`：执行失败时返回错误。

### 16.29 BindDID（UserRepo 方法）

- **函数用途**
  - 为指定地址用户绑定 DID 字段。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `address`（`string`）：用户地址。
  - `did`（`string`）：DID 字符串。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.30 GetByAddress（UserRepo 方法）

- **函数用途**
  - 按地址查询用户（大小写不敏感）。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `address`（`string`）：用户地址。

- **返回参数说明**
  - `*models.User`：用户记录。
  - `error`：未找到或查询失败时返回错误。

### 16.31 NewCredentialRepo

- **函数用途**
  - 创建可验证凭证仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*CredentialRepo`：CredentialRepo 实例。

### 16.32 Insert（CredentialRepo 方法）

- **函数用途**
  - 插入一条用户凭证记录。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `c`（`Credential`）：凭证实体，含 UserAddress、CredentialType、VCJSON、ExpiresAt。

- **返回参数说明**
  - `int64`：新插入记录的 ID。
  - `error`：插入失败时返回错误。

### 16.33 ListByUser（CredentialRepo 方法）

- **函数用途**
  - 列出用户未撤销且未过期的凭证。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `address`（`string`）：用户地址。

- **返回参数说明**
  - `[]Credential`：凭证列表。
  - `error`：查询失败时返回错误。

### 16.34 HasValidType（CredentialRepo 方法）

- **函数用途**
  - 检查用户是否持有指定类型且有效（未撤销、未过期）的凭证。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `address`（`string`）：用户地址。
  - `credType`（`string`）：凭证类型名。

- **返回参数说明**
  - `bool`：存在有效凭证为 `true`。
  - `error`：查询失败时返回错误。

### 16.35 NewComplianceRepo

- **函数用途**
  - 创建合规日志仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*ComplianceRepo`：ComplianceRepo 实例。

### 16.36 LogGeo（ComplianceRepo 方法）

- **函数用途**
  - 记录地理围栏访问日志到 `geo_access_log`。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `ip`（`string`）：客户端 IP。
  - `country`（`string`）：国家代码。
  - `path`（`string`）：请求路径。
  - `allowed`（`bool`）：是否允许访问。

- **返回参数说明**
  - `error`：插入失败时返回错误。

### 16.37 LogKYC（ComplianceRepo 方法）

- **函数用途**
  - 记录 KYC Webhook 事件到 `kyc_events`。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `externalID`（`string`）：KYC 外部 ID。
  - `userAddr`（`string`）：用户地址。
  - `status`（`string`）：KYC 状态。
  - `raw`（`[]byte`）：原始 JSON 请求体。

- **返回参数说明**
  - `error`：插入失败时返回错误。

### 16.38 NewStatsRepo

- **函数用途**
  - 创建平台统计仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*StatsRepo`：StatsRepo 实例。

### 16.39 Platform（StatsRepo 方法）

- **函数用途**
  - 聚合平台级统计：交易数、成交量、估算手续费、活跃用户、开放市场数、TVL 近似值。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。

- **返回参数说明**
  - `*PlatformStats`：统计结构体，各字段为 int64 或 numeric 字符串。
  - `error`：查询失败时返回错误。

### 16.40 UpdateMarketPool（StatsRepo 方法）

- **函数用途**
  - 更新市场的 reserve_yes、reserve_no 与 price_yes_bps 字段。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `yes`（`string`）：YES 储备 numeric 字符串。
  - `no`（`string`）：NO 储备 numeric 字符串。
  - `priceBps`（`int`）：YES 价格基点。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.41 NewOracleJobRepo

- **函数用途**
  - 创建 Oracle 作业仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*OracleJobRepo`：OracleJobRepo 实例。

### 16.42 Create（OracleJobRepo 方法）

- **函数用途**
  - 为市场创建 pending 状态的 Oracle 作业，指定最早执行时间。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `matchID`（`*int64`）：关联比赛 ID，可为 nil。
  - `executeAfter`（`time.Time`）：冷却期后可执行的时间点。

- **返回参数说明**
  - `int64`：新作业 ID。
  - `error`：插入失败时返回错误。

### 16.43 HasActiveForMarket（OracleJobRepo 方法）

- **函数用途**
  - 检查市场是否已有 pending、submitted 或 manual_review 状态的活跃作业。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `marketID`（`int64`）：市场 ID。

- **返回参数说明**
  - `bool`：存在活跃作业为 `true`。
  - `error`：查询失败时返回错误。

### 16.44 ListDue（OracleJobRepo 方法）

- **函数用途**
  - 列出 status=pending 且 execute_after ≤ now 的到期作业，JOIN 市场地址与 question。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `now`（`time.Time`）：当前时间。

- **返回参数说明**
  - `[]OracleJob`：到期作业列表。
  - `error`：查询失败时返回错误。

### 16.45 ListAll（OracleJobRepo 方法）

- **函数用途**
  - 分页列出 Oracle 作业，可按 status 过滤，按 updated_at 降序。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。
  - `status`（`string`）：状态过滤；空表示全部。
  - `limit`（`int`）：最大条数；≤0 默认 50。

- **返回参数说明**
  - `[]OracleJob`：作业列表。
  - `error`：查询失败时返回错误。

### 16.46 UpdateStatus（OracleJobRepo 方法）

- **函数用途**
  - 更新作业状态及可选字段（主/次源比分、proposed_outcome、tx_hash、error_message）。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `id`（`int64`）：作业 ID。
  - `status`（`string`）：新状态。
  - `fields`（`map[string]interface{}`）：可选更新字段，键名包括 `primary_home`、`primary_away`、`secondary_home`、`secondary_away`、`proposed_outcome`、`tx_hash`、`error_message`；值为 nil 的字段保留原值（COALESCE）。

- **返回参数说明**
  - `error`：更新失败时返回错误。

### 16.47 scanOracleJobRow

- **函数用途**
  - 从数据库行扫描 OracleJob 结构体（含可选 tx_hash、error_message 指针）。

- **函数参数说明**
  - `rows`：实现 `Scan(...)` 的行源。

- **返回参数说明**
  - `OracleJob`：扫描后的作业。
  - `error`：Scan 失败时返回错误。

### 16.48 NewIndexerStateRepo

- **函数用途**
  - 创建索引器状态仓储。

- **函数参数说明**
  - `pool`（`*pgxpool.Pool`）：数据库连接池。

- **返回参数说明**
  - `*IndexerStateRepo`：IndexerStateRepo 实例。

### 16.49 GetLastBlock（IndexerStateRepo 方法）

- **函数用途**
  - 读取索引器已处理的最后区块号（id=1 单行）。

- **函数参数说明**
  - `ctx`（`context.Context`）：查询上下文。

- **返回参数说明**
  - `uint64`：最后区块号。
  - `error`：查询失败时返回错误。

### 16.50 SetLastBlock（IndexerStateRepo 方法）

- **函数用途**
  - 更新索引器 last_block 与可选 factory_address。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `block`（`uint64`）：已处理到的区块号。
  - `factory`（`string`）：工厂合约地址；空时保留原 factory_address。

- **返回参数说明**
  - `error`：更新失败时返回错误。

---

## 17. internal/handler

### 17.1 RegisterRoutes（API 方法）

- **函数用途**
  - 在 Chi Router 上注册全部 REST/SSE/管理端路由及 JWT、Admin 中间件分组。

- **函数参数说明**
  - `r`（`chi.Router`）：路由实例。

- **返回参数说明**
  - 无返回值。

### 17.2 writeJSON

- **函数用途**
  - 写入 JSON HTTP 响应，设置 Content-Type 与状态码。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `status`（`int`）：HTTP 状态码。
  - `v`（`interface{}`）：可 JSON 序列化的响应体。

- **返回参数说明**
  - 无返回值。

### 17.3 writeError

- **函数用途**
  - 返回统一格式的 JSON 错误体 `{"error": msg}`。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `status`（`int`）：HTTP 状态码。
  - `msg`（`string`）：错误消息。

- **返回参数说明**
  - 无返回值。

### 17.4 queryInt

- **函数用途**
  - 从 URL 查询参数解析整数，缺失或无效时返回默认值。

- **函数参数说明**
  - `r`（`*http.Request`）：HTTP 请求。
  - `key`（`string`）：查询参数名。
  - `def`（`int`）：默认值。

- **返回参数说明**
  - `int`：解析结果或默认值。

### 17.5 parseID

- **函数用途**
  - 从 Chi 路由参数 `{id}` 解析 int64 主键。

- **函数参数说明**
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - `int64`：解析后的 ID。
  - `error`：非数字时返回解析错误。

### 17.6 RegisterRoutes（Health 方法）

- **函数用途**
  - 注册 `/health` 与 `/ready` 健康检查路由。

- **函数参数说明**
  - `mux`：实现 `Get(string, http.HandlerFunc)` 的路由器。

- **返回参数说明**
  - 无返回值。

### 17.7 health（Health 方法）

- **函数用途**
  - 存活探针，始终返回 `{"status":"ok"}`。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `_`（`*http.Request`）：请求（未使用）。

- **返回参数说明**
  - 无返回值；HTTP 200 JSON。

### 17.8 ready（Health 方法）

- **函数用途**
  - 就绪探针，检测 DB Ping、Redis Ping（降级不阻断）、RPC 状态；DB 失败返回 503。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；JSON 含 `status`、`db_ok`、`redis_ok`、`rpc_ok`、`chain_id` 等字段。

### 17.9 siweAuth（API 方法）

- **函数用途**
  - `POST /auth/siwe`：验证 SIWE 消息与签名，Upsert 用户并签发 24 小时 JWT。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：请求体 JSON 含 `message`、`signature`。

- **返回参数说明**
  - 无返回值；成功 200 返回 `token` 与 `user`；失败 400/401/500。

### 17.10 bindDID（API 方法）

- **函数用途**
  - `POST /users/bind-did`（需 JWT）：校验 DID 格式并绑定到当前用户。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：请求体 JSON 含 `did`、`signature`；地址来自 Context。

- **返回参数说明**
  - 无返回值；成功 200 返回更新后的 `user`。

### 17.11 myPositions（API 方法）

- **函数用途**
  - `GET /me/positions`（需 JWT）：返回当前用户全部持仓。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"items": [...]}`。

### 17.12 listMatches（API 方法）

- **函数用途**
  - `GET /matches`：分页查询比赛列表，支持 `status`、`limit`、`offset` 查询参数。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"items": [...]}`。

### 17.13 getMatch（API 方法）

- **函数用途**
  - `GET /matches/{id}`：返回比赛详情及关联市场列表。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`。

- **返回参数说明**
  - 无返回值；成功 200 返回 `match` 与 `markets`；无效 ID 400，未找到 404。

### 17.14 listMarkets（API 方法）

- **函数用途**
  - `GET /markets`：分页查询市场，响应附带 `collateral_address` 与 `chain_id`。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：支持 `status`、`limit`、`offset` 查询参数。

- **返回参数说明**
  - 无返回值；成功 200 JSON。

### 17.15 getMarket（API 方法）

- **函数用途**
  - `GET /markets/{id}`：返回市场详情及 VC 访问门禁 `access` 信息。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`；若市场 requires_vc 会检查 JWT 用户凭证。

- **返回参数说明**
  - 无返回值；成功 200 返回 `market`、`collateral_address`、`chain_id`、`access`。

### 17.16 listOracleJobs（API 方法）

- **函数用途**
  - `GET /admin/oracle-jobs`（需 Admin）：列出 Oracle 作业，可选 `status` 过滤。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；成功 200 返回最多 100 条 `items`。

### 17.17 voidMarket（API 方法）

- **函数用途**
  - `POST /admin/markets/{id}/void`（需 Admin）：链上 void 市场并更新 DB，关联比赛设为 CANCELLED。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`；可选 JSON body `reason`（当前未持久化）。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"status":"void"}`。

### 17.18 registerMarket（API 方法）

- **函数用途**
  - `POST /admin/markets`（需 Admin）：更新比赛关联市场的合规与结算规则。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：JSON body 含 `match_id`、`market_address`（未直接使用）、`question`（未直接使用）、`requires_vc`、`restricted_region`、`resolution_rule`。

- **返回参数说明**
  - 无返回值；成功 201 返回 `{"status":"registered"}`。

### 17.19 retryOracleJob（API 方法）

- **函数用途**
  - `POST /admin/oracle-jobs/{id}/retry`（需 Admin）：将作业重置为 pending 并清空 error_message。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"status":"pending"}`。

### 17.20 platformStats（API 方法）

- **函数用途**
  - `GET /stats/platform`：返回平台聚合统计数据。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；成功 200 返回 `PlatformStats` JSON。

### 17.21 marketPool（API 方法）

- **函数用途**
  - `GET /markets/{id}/pool`：返回市场 CPMM 池子与价格相关字段。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`。

- **返回参数说明**
  - 无返回值；成功 200 返回 market_id、reserve、price_bps、fee_bps 等。

### 17.22 marketOrderbook（API 方法）

- **函数用途**
  - `GET /markets/{id}/orderbook`：返回二元市场 CPMM 合成盘口快照（yes/no bids）。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：路径参数 `id`。

- **返回参数说明**
  - 无返回值；成功 200 返回 `bids` 数组与说明 `note`。

### 17.23 parseBps

- **函数用途**
  - 解析价格基点字符串；空或 0 时默认 5000（50%）。

- **函数参数说明**
  - `s`（`string`）：基点 numeric 字符串。

- **返回参数说明**
  - `int`：基点整数。

### 17.24 coalesceReserve

- **函数用途**
  - 选取首选储备金字符串；为空或 `"0"` 时使用 fallback（如 yes_pool/no_pool）。

- **函数参数说明**
  - `primary`（`string`）：首选值（如 reserve_yes）。
  - `fallback`（`string`）：备用值（如 yes_pool）。

- **返回参数说明**
  - `string`：最终使用的储备字符串。

### 17.25 issueCredential（API 方法）

- **函数用途**
  - `POST /credentials/issue`（需 Admin）：为用户签发 VC 并写入 credentials 表。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：JSON body 含 `address`、`credential_type`、`claims`、`ttl_hours`（可选）。

- **返回参数说明**
  - 无返回值；成功 201 返回 `id` 与 `vc`；缺字段 400。

### 17.26 myCredentials（API 方法）

- **函数用途**
  - `GET /users/me/credentials`（需 JWT）：列出当前用户有效凭证。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；成功 200 返回 `items`。

### 17.27 verifyVC（API 方法）

- **函数用途**
  - `POST /auth/verify-vc`：校验 VC 签名与过期，可选校验 region 与 credentialSubject 一致。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：JSON body 含 `vc_json`、`credential_type`（未强制校验类型）、`region`（可选）。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"valid":true}`；无效 401，region 不符 403。

### 17.28 complianceRestricted（API 方法）

- **函数用途**
  - `GET /compliance/restricted`：返回当前请求推断国家是否受限及合规开关。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：从 `CF-IPCountry` 或 `X-Country-Code` 读国家。

- **返回参数说明**
  - 无返回值；成功 200 返回 `country`、`restricted`、`compliance_required`、`environment`。

### 17.29 kycWebhook（API 方法）

- **函数用途**
  - `POST /kyc/webhook`：接收 KYC 回调，校验 HMAC 签名（若配置 secret），写合规日志；approved 时尝试签发 KYC VC。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：JSON body 含 `external_id`、`user_address`、`status`；Header `X-KYC-Signature` 用于验签。

- **返回参数说明**
  - 无返回值；成功 200 返回 `{"ok":"true"}`。

### 17.30 streamScores（API 方法）

- **函数用途**
  - `GET /events/scores`：SSE 流，每 5 秒推送最新比赛列表 JSON。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：须支持 Flusher 的 ResponseWriter。
  - `r`（`*http.Request`）：HTTP 请求；客户端断开时 Context 取消。

- **返回参数说明**
  - 无返回值；持续写入 `data: {...}\n\n` SSE 事件。

### 17.31 prometheusMetrics（API 方法）

- **函数用途**
  - `GET /metrics`：以 Prometheus 文本格式输出 Oracle 作业各状态计数。

- **函数参数说明**
  - `w`（`http.ResponseWriter`）：响应写入器。
  - `r`（`*http.Request`）：HTTP 请求。

- **返回参数说明**
  - 无返回值；Content-Type `text/plain`，输出 pending/manual_review/confirmed/failed 计数。

---

## 18. internal/server

### 18.1 New

- **函数用途**
  - 组装 Chi HTTP 服务器：挂载 RequestID/Logger/Recoverer、限流、地理围栏、CORS、健康检查与 API 路由，注入各 Repository 与 VC/Oracle 依赖。

- **函数参数说明**
  - `deps`（`Dependencies`）：服务器依赖。
    - `Port`（`string`）：监听端口（不含冒号）。
    - `Cfg`（`*config.Config`）：应用配置。
    - `DB`（`*pgxpool.Pool`）：PostgreSQL 连接池。
    - `Redis`（`*redis.Client`）：Redis 客户端，可为 nil。
    - `Chain`（`*blockchain.Client`）：RPC 健康客户端。
    - `OracleChain`（`*blockchain.OracleClient`）：管理端链上操作客户端，可为 nil。

- **返回参数说明**
  - `*Server`：封装 `http.Server` 的服务器实例；WriteTimeout 为 0 以支持 SSE。

### 18.2 ListenAndServe（Server 方法）

- **函数用途**
  - 启动 HTTP 监听，阻塞直到错误或 Shutdown。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `error`：正常关闭外的一般为 `http.ErrServerClosed` 或其他监听错误。

### 18.3 Shutdown（Server 方法）

- **函数用途**
  - 优雅关闭 HTTP 服务器。

- **函数参数说明**
  - `ctx`（`context.Context`）：关闭超时上下文。

- **返回参数说明**
  - `error`：Shutdown 失败时返回错误。

### 18.4 Addr（Server 方法）

- **函数用途**
  - 返回 HTTP 服务器监听地址（如 `:8080`）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `string`：Listen 地址字符串。

### 18.5 String（Server 方法）

- **函数用途**
  - 返回便于日志打印的本地访问 URL（如 `http://localhost:8080`）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `string`：完整 URL 字符串。

---

## 19. internal/indexer

### 19.1 New

- **函数用途**
  - 创建链上索引器：Dial RPC、加载 MarketFactory 与 PredictionMarket ABI。

- **函数参数说明**
  - `ctx`（`context.Context`）：Dial 上下文。
  - `cfg`（`*config.Config`）：配置，含 RPC URL、Factory 地址、ChainID 等。
  - `matches`（`*repository.MatchRepo`）：比赛仓储。
  - `markets`（`*repository.MarketRepo`）：市场仓储。
  - `positions`（`*repository.PositionRepo`）：持仓仓储。
  - `state`（`*repository.IndexerStateRepo`）：索引进度仓储。

- **返回参数说明**
  - `*Indexer`：索引器实例。
  - `error`：Dial 或 ABI 加载失败时返回错误。

### 19.2 loadABI

- **函数用途**
  - 从 `pkg/contracts/{name}.json` 加载指定合约 ABI。

- **函数参数说明**
  - `name`（`string`）：合约文件名（不含 `.json`），如 `MarketFactory`。

- **返回参数说明**
  - `abi.ABI`：解析后的 ABI。
  - `error`：文件或 JSON 错误时返回错误。

### 19.3 Run（Indexer 方法）

- **函数用途**
  - 主循环：若未配置 Factory 地址则跳过；否则按 `IndexerPollSeconds` 定时 poll 链上事件。

- **函数参数说明**
  - `ctx`（`context.Context`）：取消上下文。

- **返回参数说明**
  - `error`：Context 取消返回 `ctx.Err()`；未配置 Factory 时直接 `nil`。

### 19.4 poll（Indexer 方法）

- **函数用途**
  - 单次索引轮询：读取 last_block、扫描新区块（最多 2000 块一批）、更新 indexer_state。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `factoryAddr`（`common.Address`）：工厂合约地址。

- **返回参数说明**
  - `error`：状态读写或扫描失败时返回错误；已追上链头返回 `nil`。

### 19.5 scanFactory（Indexer 方法）

- **函数用途**
  - 在区块范围内 FilterLogs 抓取 `MarketCreated` 事件并逐条处理。

- **函数参数说明**
  - `ctx`（`context.Context`）：Filter 上下文。
  - `factory`（`common.Address`）：工厂地址。
  - `from`（`uint64`）：起始区块（含）。
  - `to`（`uint64`）：结束区块（含）。

- **返回参数说明**
  - `error`：FilterLogs 失败时返回错误；单条 handle 失败仅打日志。

### 19.6 handleMarketCreated（Indexer 方法）

- **函数用途**
  - 解析 MarketCreated 日志，关联 external match ID，InsertFromChain 写入 markets 表。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `lg`（`types.Log`）：以太坊事件日志。

- **返回参数说明**
  - `error`：Unpack 或 Insert 失败时返回错误；Topics 不足 4 个时返回 `nil`。

### 19.7 matchRefToExternal

- **函数用途**
  - 将链上 matchRef（Keccak256 哈希 hex）映射为已知 external_id；未知则原样返回 hex。

- **函数参数说明**
  - `matchRefHex`（`string`）：matchRef 十六进制字符串。

- **返回参数说明**
  - `string`：external_id 或原始 matchRefHex。

### 19.8 scanMarkets（Indexer 方法）

- **函数用途**
  - 对 DB 中所有市场（最多 500）在新区块范围内扫描 Bought、Resolved、Claimed 事件。

- **函数参数说明**
  - `ctx`（`context.Context`）：Filter 上下文。
  - `from`（`uint64`）：起始区块。
  - `to`（`uint64`）：结束区块。

- **返回参数说明**
  - `error`：List 或 FilterLogs 失败时返回错误。

### 19.9 handleBought（Indexer 方法）

- **函数用途**
  - 处理 Bought 事件：写入 trades 表并累加 positions。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：DB 市场 ID。
  - `lg`（`types.Log`）：事件日志，Topics[1] 为用户地址。

- **返回参数说明**
  - `error`：Unpack、InsertTrade 或 AddTrade 失败时返回错误。

### 19.10 handleResolved（Indexer 方法）

- **函数用途**
  - 处理 Resolved 事件，更新市场为 RESOLVED 及 winning_outcome。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketAddress`（`string`）：市场合约地址。
  - `lg`（`types.Log`）：事件日志。

- **返回参数说明**
  - `error`：Unpack 或 UpdateResolved 失败时返回错误。

### 19.11 handleClaimed（Indexer 方法）

- **函数用途**
  - 处理 Claimed 事件，将用户持仓标记为 claimed。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `marketID`（`int64`）：市场 ID。
  - `lg`（`types.Log`）：事件日志，Topics[1] 为用户地址。

- **返回参数说明**
  - `error`：SetClaimed 失败时返回错误。

---

## 20. internal/oracle

### 20.1 NewWorker

- **函数用途**
  - 创建 Oracle Worker，注入配置、仓储、双源比分 Provider、链上客户端与告警器。

- **函数参数说明**
  - `cfg`（`*config.Config`）：应用配置。
  - `jobs`（`*repository.OracleJobRepo`）：Oracle 作业仓储。
  - `matches`（`*repository.MatchRepo`）：比赛仓储。
  - `markets`（`*repository.MarketRepo`）：市场仓储。
  - `dual`（`*wcprovider.DualProvider`）：双源比分 Provider。
  - `chain`（`*blockchain.OracleClient`）：链上 Oracle 客户端，可为 nil。
  - `alerts`（`*alert.Notifier`）：告警通知器。

- **返回参数说明**
  - `*Worker`：Worker 实例。

### 20.2 Run（Worker 方法）

- **函数用途**
  - 主循环：按 `OraclePollSeconds` 定时执行 tick，直到 Context 取消。

- **函数参数说明**
  - `ctx`（`context.Context`）：取消上下文。

- **返回参数说明**
  - `error`：Context 取消时返回 `ctx.Err()`。

### 20.3 tick（Worker 方法）

- **函数用途**
  - 单次调度：先为已结束比赛入队 Oracle 作业，再处理到期作业。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。

- **返回参数说明**
  - `error`：enqueue 或 processDue 失败时返回错误。

### 20.4 enqueueFinished（Worker 方法）

- **函数用途**
  - 扫描 FINISHED 状态比赛，为其 OPEN 市场创建 pending Oracle 作业（冷却期后执行），并将比赛设为 ORACLE_PENDING。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。

- **返回参数说明**
  - `error`：List 或 Create 批量失败时返回错误。

### 20.5 processDue（Worker 方法）

- **函数用途**
  - 列出到期 pending 作业并逐个 processJob；chain 为 nil 时跳过。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。

- **返回参数说明**
  - `error`：ListDue 失败时返回错误；单 job 失败仅打日志。

### 20.6 processJob（Worker 方法）

- **函数用途**
  - 处理单个 Oracle 作业：双源比分比对、按 resolution_rule 计算 outcome、链上 resolve（立即或 request+confirm 流程）、更新 DB 与比赛状态。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `job`（`repository.OracleJob`）：待处理作业，含 MarketID、MatchID、MarketAddress 等。

- **返回参数说明**
  - `error`：比对失败、链上失败或 DB 更新失败时可能返回 error；双源不一致时标记 manual_review 并返回 `nil`。

### 20.7 merge

- **函数用途**
  - 将 extra map 的键值合并到 base map（原地修改 base）。

- **函数参数说明**
  - `base`（`map[string]interface{}`）：基础字段 map。
  - `extra`（`map[string]interface{}`）：待合并的额外字段。

- **返回参数说明**
  - `map[string]interface{}`：合并后的 base（同一引用）。

---

## 21. internal/wcprovider

### 21.1 NewMock

- **函数用途**
  - 创建 Mock 世界杯数据 Provider，指向本地 JSON 文件路径。

- **函数参数说明**
  - `path`（`string`）：Mock 比赛 JSON 文件路径。

- **返回参数说明**
  - `*MockProvider`：MockProvider 实例。

### 21.2 Sync（MockProvider 方法）

- **函数用途**
  - 读取 JSON 文件并 Upsert 全部比赛到数据库。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `repo`（`*repository.MatchRepo`）：比赛仓储。

- **返回参数说明**
  - `int`：成功同步条数。
  - `error`：读文件、JSON 解析、时间解析或 Upsert 失败时返回错误。

### 21.3 NewDual

- **函数用途**
  - 创建双源比分 Provider，指定主/次 JSON 文件路径。

- **函数参数说明**
  - `primary`（`string`）：主数据源 JSON 路径。
  - `secondary`（`string`）：次数据源 JSON 路径。

- **返回参数说明**
  - `*DualProvider`：DualProvider 实例。

### 21.4 load（DualProvider 方法）

- **函数用途**
  - 加载 JSON 文件中含完整比分的比赛，构建 external_id → ScoreSnapshot map。

- **函数参数说明**
  - `path`（`string`）：JSON 文件路径。

- **返回参数说明**
  - `map[string]ScoreSnapshot`：external_id 到比分快照的映射。
  - `error`：读文件或 JSON 错误时返回错误。

### 21.5 CompareScores（DualProvider 方法）

- **函数用途**
  - 对比主/次数据源中同一 external_id 的主客比分是否一致。

- **函数参数说明**
  - `externalID`（`string`）：比赛外部 ID。

- **返回参数说明**
  - `primary`（`ScoreSnapshot`）：主源比分快照。
  - `secondary`（`ScoreSnapshot`）：次源比分快照。
  - `match`（`bool`）：两边均有数据且主客比分完全一致为 `true`；任一侧缺失为 `false`。
  - `err`（`error`）：load 文件失败时返回错误。

### 21.6 SyncPrimary（DualProvider 方法）

- **函数用途**
  - 仅同步主数据源 JSON 到数据库（委托 MockProvider.Sync）。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `repo`（`*repository.MatchRepo`）：比赛仓储。

- **返回参数说明**
  - `int`：同步条数。
  - `error`：同步失败时返回错误。

### 21.7 OutcomeFromRule

- **函数用途**
  - 根据结算规则与主客比分计算二元市场 outcome（0 或 1）。

- **函数参数说明**
  - `rule`（`string`）：规则名，`OVER_25` 表示总进球大于 2 为 outcome 0；默认 `HOME_WIN` 表示主队胜为 0。
  - `home`（`int`）：主队得分。
  - `away`（`int`）：客队得分。

- **返回参数说明**
  - `int`：outcome 索引，0 或 1。

### 21.8 FinishMatch（DualProvider 方法）

- **函数用途**
  - 将指定 external_id 比赛设为 FINISHED 并写入最终比分后 Upsert。

- **函数参数说明**
  - `ctx`（`context.Context`）：执行上下文。
  - `repo`（`*repository.MatchRepo`）：比赛仓储。
  - `externalID`（`string`）：比赛外部 ID。
  - `home`（`int`）：主队最终比分。
  - `away`（`int`）：客队最终比分。

- **返回参数说明**
  - `error`：GetByExternalID 或 Upsert 失败时返回错误。

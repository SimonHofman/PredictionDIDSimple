# Frontend 函数说明文档

本文档按模块整理 `frontend/src/` 目录下 JavaScript/JSX 源码中的函数，包含函数用途、参数说明与返回值说明。React 组件默认导出函数以 `props` 作为入参、以 JSX 作为渲染返回值；组件内部定义的辅助函数单独列出。

---

## 1. main.jsx

应用入口，挂载 React 根节点并包裹路由、Web3、国际化 Provider。

### 1.1 （入口脚本）

- **函数用途**
  - 调用 `ReactDOM.createRoot(...).render(...)` 启动应用，无自定义命名函数。

- **函数参数说明**
  - 无自定义函数参数。

- **返回参数说明**
  - 无返回值。

---

## 2. config.js

环境变量配置模块，导出 `config` 常量对象（非函数），供 API、SIWE、链 ID 等读取。

---

## 3. wagmi.js

Wagmi/Viem Web3 配置模块，通过 `defineChain` 与 `createConfig` 导出 `hardhatLocal` 链定义与 `wagmiConfig` 实例（无项目内自定义命名函数）。

---

## 4. services/api.js

后端 REST API 封装层，统一管理 JWT、请求头与错误处理。

### 4.1 getToken

- **函数用途**
  - 从 `localStorage` 读取 SIWE 登录后的 JWT 令牌。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `string | null`：JWT 字符串；未登录时为 `null`。

### 4.2 setToken

- **函数用途**
  - 将 JWT 写入 `localStorage`（键名 `prediction_jwt`）。

- **函数参数说明**
  - `token`（`string`）：待保存的 JWT。

- **返回参数说明**
  - 无返回值。

### 4.3 clearToken

- **函数用途**
  - 清除 `localStorage` 中的 JWT，用于登出。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值。

### 4.4 request

- **函数用途**
  - 内部通用 HTTP 请求：拼接 `config.apiUrl`、自动附加 JSON Content-Type 与 Bearer JWT，解析 JSON 响应，非 2xx 时抛出 Error。

- **函数参数说明**
  - `path`（`string`）：API 路径，如 `/matches`。
  - `options`（`object`，可选）：传给 `fetch` 的选项，可含 `method`、`body`、`headers` 等；默认 `{}`。

- **返回参数说明**
  - `Promise<object>`：解析后的 JSON 响应体。
  - 失败时 reject，`Error.message` 为 `data.error` 或 `HTTP {status}`。

### 4.5 getHealth

- **函数用途**
  - 调用 `GET /health` 检测 API 存活。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<object>`：如 `{ status: "ok" }`。

### 4.6 getReady

- **函数用途**
  - 调用 `GET /ready` 就绪探针，不经过 `request` 封装，返回原始 ok 状态。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<{ ok: boolean, data: object }>`：`ok` 为 HTTP 是否成功，`data` 为 JSON 体。

### 4.7 listMatches

- **函数用途**
  - 分页查询比赛列表 `GET /matches`。

- **函数参数说明**
  - `params`（`object`，可选）：查询参数，如 `{ status, limit, offset }`；默认 `{}`。

- **返回参数说明**
  - `Promise<object>`：含 `items` 数组的响应。

### 4.8 getMatch

- **函数用途**
  - 查询单场比赛详情 `GET /matches/{id}`。

- **函数参数说明**
  - `id`（`string | number`）：比赛 ID。

- **返回参数说明**
  - `Promise<object>`：含 `match` 与 `markets` 等字段。

### 4.9 listMarkets

- **函数用途**
  - 分页查询市场列表 `GET /markets`。

- **函数参数说明**
  - `params`（`object`，可选）：如 `{ status, limit, offset }`；默认 `{}`。

- **返回参数说明**
  - `Promise<object>`：含 `items`、`collateral_address`、`chain_id` 等。

### 4.10 getMarket

- **函数用途**
  - 查询单个市场详情 `GET /markets/{id}`。

- **函数参数说明**
  - `id`（`string | number`）：市场 ID。

- **返回参数说明**
  - `Promise<object>`：含 `market`、`access`、`collateral_address`、`chain_id`。

### 4.11 siweAuth

- **函数用途**
  - SIWE 登录 `POST /auth/siwe`，提交消息与签名换取 JWT。

- **函数参数说明**
  - `message`（`string`）：SIWE 格式化消息文本。
  - `signature`（`string`）：钱包签名 hex。

- **返回参数说明**
  - `Promise<object>`：含 `token` 与 `user` 对象。

### 4.12 bindDid

- **函数用途**
  - 绑定 DID `POST /users/bind-did`（需 JWT）。

- **函数参数说明**
  - `did`（`string`）：DID 字符串，如 `did:pkh:eip155:31337:0x...`。
  - `signature`（`string`）：绑定签名（后端 MVP 可选校验）。

- **返回参数说明**
  - `Promise<object>`：含更新后的 `user`。

### 4.13 myPositions

- **函数用途**
  - 查询当前用户持仓 `GET /me/positions`（需 JWT）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<object>`：含 `items` 持仓数组。

### 4.14 myCredentials

- **函数用途**
  - 查询当前用户可验证凭证 `GET /users/me/credentials`（需 JWT）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<object>`：含 `items` 凭证数组。

### 4.15 verifyVC

- **函数用途**
  - 校验 VC `POST /auth/verify-vc`。

- **函数参数说明**
  - `vcJson`（`object | string`）：VC JSON 对象或字符串。
  - `credentialType`（`string`）：凭证类型名。
  - `region`（`string`）：可选地区码，用于 region 校验。

- **返回参数说明**
  - `Promise<object>`：成功时 `{ valid: true }`。

### 4.16 adminHeaders

- **函数用途**
  - 内部函数，从 `sessionStorage.admin_key` 构造 `X-Admin-Key` 请求头。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `object`：`{ 'X-Admin-Key': string }`。

### 4.17 adminListOracleJobs

- **函数用途**
  - 管理员列出 Oracle 作业 `GET /admin/oracle-jobs`。

- **函数参数说明**
  - `status`（`string`，可选）：状态过滤；默认 `''` 表示全部。

- **返回参数说明**
  - `Promise<object>`：含 `items` 作业数组。

### 4.18 adminRetryOracleJob

- **函数用途**
  - 管理员重试 Oracle 作业 `POST /admin/oracle-jobs/{id}/retry`。

- **函数参数说明**
  - `id`（`string | number`）：作业 ID。

- **返回参数说明**
  - `Promise<object>`：如 `{ status: "pending" }`。

### 4.19 adminRegisterMarket

- **函数用途**
  - 管理员登记/更新市场合规规则 `POST /admin/markets`。

- **函数参数说明**
  - `body`（`object`）：请求体，含 `match_id`、`requires_vc`、`restricted_region`、`resolution_rule` 等。

- **返回参数说明**
  - `Promise<object>`：如 `{ status: "registered" }`。

### 4.20 adminVoidMarket

- **函数用途**
  - 管理员作废市场 `POST /admin/markets/{id}/void`。

- **函数参数说明**
  - `id`（`string | number`）：市场 ID。

- **返回参数说明**
  - `Promise<object>`：如 `{ status: "void" }`。

### 4.21 getCompliance

- **函数用途**
  - 查询合规/地理限制状态 `GET /compliance/restricted`。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<object>`：含 `country`、`restricted`、`compliance_required`、`environment`。

### 4.22 getPlatformStats

- **函数用途**
  - 查询平台统计 `GET /stats/platform`。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<object>`：含 `trade_count`、`trade_volume`、`fees_collected`、`active_users`、`open_markets`、`tvl_approx`。

### 4.23 getMarketPool

- **函数用途**
  - 查询市场池子快照 `GET /markets/{id}/pool`。

- **函数参数说明**
  - `id`（`string | number`）：市场 ID。

- **返回参数说明**
  - `Promise<object>`：含 `market_id`、`reserve_yes`、`reserve_no`、`price_yes_bps`、`fee_bps` 等。

### 4.24 getMarketOrderbook

- **函数用途**
  - 查询市场合成盘口 `GET /markets/{id}/orderbook`。

- **函数参数说明**
  - `id`（`string | number`）：市场 ID。

- **返回参数说明**
  - `Promise<object>`：含 `bids` 数组与 `note`。

---

## 5. services/contracts.js

链上合约交互封装，基于 Wagmi `readContract` / `writeContract` 与 viem 单位转换。

### 5.1 parseUsdc

- **函数用途**
  - 将人类可读 USDC/mUSDC 数量字符串转为链上最小单位 `bigint`（6 位小数）。

- **函数参数说明**
  - `amount`（`string | number`）：可读金额，如 `"10"`。

- **返回参数说明**
  - `bigint`：6 位精度的 token 数量。

### 5.2 formatUsdc

- **函数用途**
  - 将链上最小单位转为可读 USDC 字符串。

- **函数参数说明**
  - `amount`（`bigint | string | number`）：链上数量；空值按 0 处理。

- **返回参数说明**
  - `string`：格式化后的十进制字符串。

### 5.3 readUsdcBalance

- **函数用途**
  - 读取 ERC20 MockUSDC 的 `balanceOf` 余额。

- **函数参数说明**
  - `address`（`string`）：持仓钱包地址（0x...）。
  - `tokenAddress`（`string`）：MockUSDC 合约地址。

- **返回参数说明**
  - `Promise<bigint>`：token 最小单位余额。

### 5.4 approveUsdc

- **函数用途**
  - 调用 MockUSDC `approve(spender, amount)` 并等待交易确认。

- **函数参数说明**
  - `tokenAddress`（`string`）：MockUSDC 合约地址。
  - `spender`（`string`）：被授权 spender 地址（通常为市场合约）。
  - `amount`（`bigint`）：授权数量（最小单位）。

- **返回参数说明**
  - `Promise<object>`：`waitForTransactionReceipt` 返回的交易回执，含 `transactionHash` 等。

### 5.5 buyOutcome

- **函数用途**
  - 在 Phase 1/2 二元 parimutuel 市场（PredictionMarket）调用 `buy(outcome, amount)`。

- **函数参数说明**
  - `marketAddress`（`string`）：市场合约地址。
  - `outcome`（`number`）：投注方向，0=Yes，1=No。
  - `amount`（`bigint`）：投注数量（最小单位）。

- **返回参数说明**
  - `Promise<object>`：交易回执。

### 5.6 claimMarket

- **函数用途**
  - 在 PredictionMarket 调用 `claim()` 领取结算或作废退款。

- **函数参数说明**
  - `marketAddress`（`string`）：市场合约地址。

- **返回参数说明**
  - `Promise<object>`：交易回执。

### 5.7 buyV3

- **函数用途**
  - 在 CPMM 二元市场（PredictionMarketV3）调用 `buy(outcome, amountIn)`。

- **函数参数说明**
  - `marketAddress`（`string`）：V3 市场合约地址。
  - `outcome`（`number`）：0=Yes，1=No。
  - `amountIn`（`bigint`）：投入总量（含 fee，最小单位）。

- **返回参数说明**
  - `Promise<object>`：交易回执。

### 5.8 buyMulti

- **函数用途**
  - 在多结果市场（MultiOutcomeMarket）调用 `buy(outcome, amount)`。

- **函数参数说明**
  - `marketAddress`（`string`）：多结果市场合约地址。
  - `outcome`（`number`）：结果索引（0 至 outcomeCount-1）。
  - `amount`（`bigint`）：投注总量（最小单位）。

- **返回参数说明**
  - `Promise<object>`：交易回执。

### 5.9 addLiquidityV3

- **函数用途**
  - 在 PredictionMarketV3 调用 `addLiquidity(amount)` 注入流动性。

- **函数参数说明**
  - `marketAddress`（`string`）：V3 市场合约地址。
  - `amount`（`bigint`）：注入 collateral 总量（最小单位）。

- **返回参数说明**
  - `Promise<object>`：交易回执。

### 5.10 readPoolStateV3

- **函数用途**
  - 读取 V3 市场 `getPoolState()`，返回储备与 YES 价格基点。

- **函数参数说明**
  - `marketAddress`（`string`）：V3 市场合约地址。

- **返回参数说明**
  - `Promise<object>`：`{ reserveYes, reserveNo, priceYesBps }`，均为 bigint 或 number。

### 5.11 readMarketStatus

- **函数用途**
  - 并行读取 PredictionMarket 的 `status`、`winningOutcome`、`yesPool`、`noPool`。

- **函数参数说明**
  - `marketAddress`（`string`）：市场合约地址。

- **返回参数说明**
  - `Promise<object>`：`{ status, winningOutcome, yesPool, noPool }`。

---

## 6. hooks/useAuth.js

SIWE 认证 React Hook，封装登录、登出与认证状态。

### 6.1 useAuth

- **函数用途**
  - 提供钱包连接、JWT 认证状态及 `login`/`logout` 方法；登录时构造 SiweMessage、签名、换 token、可选 bindDid 并刷新页面。

- **函数参数说明**
  - 无参数（React Hook）。

- **返回参数说明**
  - `object`：
    - `isConnected`（`boolean`）：钱包是否已连接。
    - `address`（`string | undefined`）：当前钱包地址。
    - `token`（`string | null`）：JWT。
    - `isAuthenticated`（`boolean`）：是否有有效 token。
    - `login`（`async function`）：触发 SIWE 登录流程。
    - `logout`（`function`）：清除 token 并刷新页面。
    - `loading`（`boolean`）：登录进行中。
    - `error`（`string | null`）：最近一次登录错误消息。

### 6.2 login（useAuth 内部）

- **函数用途**
  - 生成 SIWE 消息、请求钱包签名、调用 `siweAuth` 与 `bindDid`，成功后 `setToken` 并 `window.location.reload()`。

- **函数参数说明**
  - 无参数；依赖 `address`、`chainId` 与 `signMessageAsync`。

- **返回参数说明**
  - `Promise<void>`：失败时设置 `error` 状态，不 throw 到外部。

### 6.3 logout（useAuth 内部）

- **函数用途**
  - 调用 `clearToken()` 并刷新页面。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值。

---

## 7. i18n/index.jsx

中英文国际化 Context。

### 7.1 I18nProvider

- **函数用途**
  - React Context Provider，提供当前语言、`t` 翻译函数与 `setLang` 切换语言；语言偏好持久化到 `localStorage.lang`。

- **函数参数说明**
  - `children`（`ReactNode`）：子组件树。

- **返回参数说明**
  - `JSX.Element`：`<I18nContext.Provider>` 包裹的 children。

### 7.2 t（I18nProvider 内部）

- **函数用途**
  - 根据当前语言从 `messages` 对象取文案，缺失 key 时返回 key 本身。

- **函数参数说明**
  - `key`（`string`）：文案键名，如 `'home'`、`'markets'`。

- **返回参数说明**
  - `string`：翻译后的文本。

### 7.3 setLanguage / setLang（I18nProvider 内部）

- **函数用途**
  - 切换语言并写入 `localStorage`，触发 re-render。

- **函数参数说明**
  - `l`（`string`）：语言代码，如 `'zh'` 或 `'en'`。

- **返回参数说明**
  - 无返回值。

### 7.4 useI18n

- **函数用途**
  - 消费 I18n Context；必须在 `I18nProvider` 内使用。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `object`：`{ lang, setLang, t }`。
  - 在 Provider 外调用时 throw `Error('useI18n outside provider')`。

---

## 8. providers/Web3Provider.jsx

Web3 与 React Query 顶层 Provider。

### 8.1 Web3Provider

- **函数用途**
  - 包裹 `WagmiProvider` 与 `QueryClientProvider`，为子树提供 wagmi 与 TanStack Query 能力。

- **函数参数说明**
  - `children`（`ReactNode`）：应用子组件。

- **返回参数说明**
  - `JSX.Element`：Provider 嵌套结构。

---

## 9. App.jsx

根路由配置组件。

### 9.1 App

- **函数用途**
  - 定义 React Router 路由表：合规门控 → Layout → 各页面及 admin 嵌套路由。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：`<Routes>` 路由树。

---

## 10. components/Layout.jsx

全局页面布局（导航栏 + Outlet + 页脚）。

### 10.1 Layout

- **函数用途**
  - 渲染顶栏导航（首页、市场、统计、流动性、我的、DID、管理）、语言切换、WalletBar 与子路由 Outlet。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：完整页面布局 DOM。

---

## 11. components/WalletBar.jsx

钱包连接与 SIWE 登录栏。

### 11.1 WalletBar

- **函数用途**
  - 显示连接钱包、地址缩写、SIWE 登录/已登录状态、断开连接；集成 `useConnect`、`useDisconnect`、`useAuth`。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：钱包操作 UI。

---

## 12. components/ComplianceWrapper.jsx

合规与地理围栏门控，包裹所有业务路由。

### 12.1 ComplianceWrapper

- **函数用途**
  - 挂载时调用 `getCompliance` 检查地区限制；受限则显示不可用页；未接受风险披露则显示确认页；通过后渲染 `<Outlet />`。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：门控页或子路由 Outlet。

---

## 13. components/TxStatus.jsx

链上交易状态展示组件。

### 13.1 TxStatus

- **函数用途**
  - 根据交易状态显示 pending/success/error 文案及可选 tx hash；无状态且无错误时不渲染。

- **函数参数说明**
  - `status`（`string | null`）：`'pending'` | `'success'` | `'error'` | `null`。
  - `error`（`string | null`）：错误消息，status 为 error 时展示。
  - `hash`（`string | null`）：交易哈希，可选展示。

- **返回参数说明**
  - `JSX.Element | null`：状态卡片或 `null`。

---

## 14. components/MarketStatusBadge.jsx

市场/比赛状态徽章。

### 14.1 MarketStatusBadge

- **函数用途**
  - 将状态码（OPEN、RESOLVED、VOID、ORACLE_PENDING 等）映射为中文标签与 CSS class。

- **函数参数说明**
  - `status`（`string`）：市场或比赛状态字符串。

- **返回参数说明**
  - `JSX.Element`：`<span className="badge ...">` 徽章。

---

## 15. components/VCCard.jsx

可验证凭证卡片展示。

### 15.1 VCCard

- **函数用途**
  - 解析 credential 的 `vc_json`，展示凭证类型、W3C type、过期时间与 credentialSubject JSON。

- **函数参数说明**
  - `credential`（`object`）：API 返回的凭证对象，含 `credential_type`、`vc_json`、`expires_at` 等。

- **返回参数说明**
  - `JSX.Element`：凭证信息卡片。

---

## 16. pages/Home.jsx

首页：API 状态与即将开始/进行中的比赛列表。

### 16.1 Home

- **函数用途**
  - 组件挂载时检测 API 健康并加载前 10 场比赛，筛选 SCHEDULED/LIVE 展示。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：首页 UI。

### 16.2 load（Home 内部）

- **函数用途**
  - 异步调用 `getHealth` 与 `listMatches`，更新 `apiStatus`、`matches`、`error` 状态。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `Promise<void>`：失败时设置 offline 与 error。

---

## 17. pages/Markets.jsx

市场列表页。

### 17.1 Markets

- **函数用途**
  - 加载最多 50 个市场，展示问题、关联比赛、Yes/No 池与状态，链接至详情页。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：市场列表 UI。

---

## 18. pages/MarketDetail.jsx

市场详情与下注页。

### 18.1 MarketDetail

- **函数用途**
  - 根据路由 `id` 加载市场详情、池状态、VC 门禁；开放且允许时提供 outcome/金额选择与 Approve+Buy。

- **函数参数说明**
  - 无 props（通过 `useParams()` 读取 `id`）。

- **返回参数说明**
  - `JSX.Element`：市场详情与下注 UI。

### 18.2 refresh（MarketDetail 内部）

- **函数用途**
  - 重新拉取 `getMarket(id)`，并按市场类型读取链上 status 或 V3 pool、API pool 快照。

- **函数参数说明**
  - 无参数；闭包依赖路由 `id` 与 `data`。

- **返回参数说明**
  - `Promise<void>`：更新 `data` 与 `pool` 状态。

### 18.3 onBuy（MarketDetail 内部）

- **函数用途**
  - 校验连接与 VC 门禁后，`approveUsdc` 再按 `market_type` 调用 `buyV3` / `buyMulti` / `buyOutcome`，更新 tx 状态并 refresh。

- **函数参数说明**
  - 无参数；使用组件 state 中的 `amount`、`outcome`、`market`、`collateral`。

- **返回参数说明**
  - `Promise<void>`：成功设置 tx success 与 hash；失败设置 error message。

---

## 19. pages/MatchDetail.jsx

比赛详情与关联市场列表。

### 19.1 MatchDetail

- **函数用途**
  - 根据路由 `id` 调用 `getMatch`，展示开球时间、状态、比分及关联市场链接。

- **函数参数说明**
  - 无 props（`useParams().id`）。

- **返回参数说明**
  - `JSX.Element`：比赛详情 UI。

---

## 20. pages/Me.jsx

用户持仓与 Claim 页。

### 20.1 Me

- **函数用途**
  - 登录后加载 `myPositions`，展示 Yes/No 持仓；RESOLVED/VOID 且未 claim 时提供 Claim 按钮。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：持仓列表 UI。

### 20.2 load（Me 内部）

- **函数用途**
  - 调用 `myPositions` 更新 `items` 或 `error`。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值（Promise 链更新 state）。

### 20.3 onClaim（Me 内部）

- **函数用途**
  - 对指定市场地址调用 `claimMarket`，更新 tx 状态并重新 load 持仓。

- **函数参数说明**
  - `marketAddress`（`string`）：市场合约地址。

- **返回参数说明**
  - `Promise<void>`：成功/失败更新 `tx` state。

---

## 21. pages/DIDProfile.jsx

DID 身份与 VC 列表页。

### 21.1 DIDProfile

- **函数用途**
  - 展示当前钱包对应的 `did:pkh:eip155:...`；登录后加载并渲染 `myCredentials` 列表。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：DID 与 VC 列表 UI。

---

## 22. pages/Stats.jsx

平台统计页。

### 22.1 Stats

- **函数用途**
  - 挂载时调用 `getPlatformStats`，格式化展示成交量、手续费、活跃用户、TVL 等。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：统计列表 UI。

---

## 23. pages/Liquidity.jsx

V3 CPMM 流动性注入页。

### 23.1 Liquidity

- **函数用途**
  - 列出 OPEN 市场，选择市场后展示 API pool 快照，支持 Approve + `addLiquidityV3`。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：流动性操作 UI。

### 23.2 onAdd（Liquidity 内部）

- **函数用途**
  - 对选中 V3 市场执行 `approveUsdc` 与 `addLiquidityV3`，刷新 pool 并更新 tx 状态字符串。

- **函数参数说明**
  - 无参数；使用 `selected`、`amount`、`markets` 与 `VITE_MOCK_USDC_ADDRESS`。

- **返回参数说明**
  - `Promise<void>`：成功 `tx='ok'`，失败 `tx=error.message`。

---

## 24. pages/admin/AdminLayout.jsx

管理后台布局与 Admin Key 提示。

### 24.1 AdminLayout

- **函数用途**
  - 检查 `sessionStorage.admin_key`，展示 Oracle/市场配置子导航与 `<Outlet />`。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：管理后台框架 UI。

---

## 25. pages/admin/OracleJobs.jsx

Oracle 作业队列管理页。

### 25.1 OracleJobs

- **函数用途**
  - 轮询（10s）`adminListOracleJobs`，展示作业状态、双源比分、错误信息；manual_review 时可重试。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：Oracle 任务列表 UI。

### 25.2 load（OracleJobs 内部）

- **函数用途**
  - 调用 `adminListOracleJobs` 更新 `items` 或 `error`。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值。

---

## 26. pages/admin/AdminMarkets.jsx

管理员市场配置与作废页。

### 26.1 AdminMarkets

- **函数用途**
  - 提供表单调用 `adminRegisterMarket` 更新 match 合规规则，以及 `adminVoidMarket` 作废指定市场 ID。

- **函数参数说明**
  - 无 props。

- **返回参数说明**
  - `JSX.Element`：市场管理表单 UI。

### 26.2 onRegister（AdminMarkets 内部）

- **函数用途**
  - 提交 match_id、requires_vc、restricted_region、resolution_rule 到 `adminRegisterMarket`。

- **函数参数说明**
  - 无参数；使用组件 state `matchId`、`requiresVC`。

- **返回参数说明**
  - `Promise<void>`：成功/失败更新 `msg` 提示。

### 26.3 onVoid（AdminMarkets 内部）

- **函数用途**
  - 对 `voidId` 调用 `adminVoidMarket`。

- **函数参数说明**
  - 无参数；使用 state `voidId`。

- **返回参数说明**
  - `Promise<void>`：成功/失败更新 `msg`。

---

## 附录：config 导出字段

`config.js` 非函数模块，导出对象字段供参考：

- `apiUrl`：后端 API 基址，默认 `http://localhost:8080`
- `chainId`：链 ID，默认 `31337`
- `apiUrlConfigured`：是否配置了 `VITE_API_URL`
- `mockUsdc`：MockUSDC 合约地址
- `marketFactory`：MarketFactory 合约地址
- `siweDomain` / `siweUri`：SIWE 域与 URI

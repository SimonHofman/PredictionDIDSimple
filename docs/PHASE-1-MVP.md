# Phase 1 — MVP（可演示的最小预测市场）

**周期建议**：约 4–6 周  
**前置依赖**：[Phase 0](./PHASE-0-基础设施.md)  
**后续 Phase**：[Phase 2](./PHASE-2-身份与预言机.md)

---

## 1. 目标

在**测试网**完成端到端闭环：浏览世界杯赛程 → 选择一场比赛的 **Yes/No 市场** → 钱包下注 → （可手动）结算 → 领取。身份采用 **SIWE + did:pkh** 轻量绑定，不做完整 VC。

## 2. 范围

### 2.1 包含

- 单结果类型：胜/负或 Yes/No（二选一）
- 一种抵押：测试 ERC20 或 Hardhat 部署的 Mock USDC
- 赛程与比分：对接**一种**世界杯/足球 API（Mock 亦可用于演示）
- 链上事件索引到 DB，列表页读 API 而非全链扫描
- 前端：连接钱包、下注、查看持仓、claim

### 2.2 不包含

- 自动 Oracle 上链（可管理员脚本 `resolve`）
- 可验证凭证（VC）、KYC
- 流动性 AMM、订单簿
- 多签 Oracle、主网、审计

---

## 3. 智能合约任务

| # | 任务 | 说明 |
|---|------|------|
| C1-1 | `MarketFactory` | `createMarket(matchRef, question, endTime, collateral)` |
| C1-2 | `PredictionMarket` | `buy(outcome, amount)`、`claim()`、`resolve(outcome)`（仅 owner/oracle） |
| C1-3 | ERC20 接口 | 抵押转入/转出、余额检查 |
| C1-4 | 事件 | `MarketCreated`、`Bought`、`Resolved`、`Claimed` |
| C1-5 | 单元测试 | 创建、下注、resolve、claim、重复 claim 失败 |
| C1-6 | 部署脚本 | 测试网 + 本地；地址写入 `.env` |

**合约接口（示意）**

```solidity
function buy(uint8 outcome, uint256 amount) external;
function resolve(uint8 winningOutcome) external onlyOracle;
function claim() external;
```

---

## 4. 后端任务（Go）

| # | 任务 | 说明 |
|---|------|------|
| B1-1 | 数据表 | `matches`、`markets`、`trades`、`positions`、`users` |
| B1-2 | WC Sync Job | 定时拉 API → 更新 `matches`（时间、比分、状态） |
| B1-3 | Indexer Worker | 订阅 `MarketCreated`/`Bought`/`Resolved` → 写 DB |
| B1-4 | REST API | 见下表 |
| B1-5 | SIWE 登录 | 验证签名 → 签发 JWT 或 session |
| B1-6 | did:pkh 绑定 | `POST /users/bind-did` 存 `address ↔ did` |

### 4.1 API 端点（MVP）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/matches` | 赛程列表（分页、状态筛选） |
| GET | `/matches/:id` | 单场详情 |
| GET | `/markets` | 市场列表（关联 match） |
| GET | `/markets/:id` | 市场详情 + 链上 id |
| GET | `/me/positions` | 需登录，当前用户持仓 |
| POST | `/auth/siwe` | SIWE 验证 |
| POST | `/users/bind-did` | DID 字符串 + 签名 |

---

## 5. 前端任务（Node.js / JS）

| # | 任务 | 说明 |
|---|------|------|
| F1-1 | 钱包连接 | wagmi 或 ethers：连接、切网、显示地址 |
| F1-2 | 首页 | 即将开始 / 进行中的比赛卡片 |
| F1-3 | 比赛详情页 | 比分、开球时间、关联市场入口 |
| F1-4 | 市场页 | 选择 Yes/No、输入金额、approve + buy |
| F1-5 | 个人中心 | 持仓列表、claim 按钮（resolved 后） |
| F1-6 | SIWE 登录流 | 与后端 session 联动（可选：仅记录 DID） |
| F1-7 | 交易反馈 | pending / success / revert 提示 |

---

## 6. 数据与集成

| 项 | 说明 |
|----|------|
| 足球 API | 申请 API-Football 等 Key；开发可用 JSON Mock |
| 映射 | `markets.chain_id` + `contract_address` + `market_id` ↔ `matches.external_id` |
| 结算（MVP） | 运维调用 Hardhat 脚本或仅 owner `resolve`；比分仅作 UI 展示 |

---

## 7. 交付物

- [x] 本地部署文档 [DEPLOYMENT-PHASE1.md](./DEPLOYMENT-PHASE1.md)
- [x] 至少 3 场测试比赛（`backend/data/mock_matches.json`）+ seed-markets 脚本
- [x] 演示流程文档：连接钱包 → 下注 → resolve → claim
- [x] OpenAPI 草稿 [api/openapi-phase1.yaml](./api/openapi-phase1.yaml)

---

## 8. 验收标准

1. 测试网完整走通一笔下注并在 `resolve` 后成功 `claim`。
2. 前端市场列表与链上状态在 1 个区块确认内与 Indexer 一致（允许可配置延迟）。
3. SIWE 登录后 `/me/positions` 能按 address 返回正确持仓。
4. 合约测试覆盖率：核心路径 ≥ 80%（团队自定阈值）。

---

## 9. 排除与延后（进入 Phase 2）

- Oracle Worker 自动上链
- VC / Ceramic / 链上 DIDRegistry
- 赛事取消自动 VOID
- 管理后台 UI

---

## 10. 里程碑建议

| 周 | 里程碑 |
|----|--------|
| 1 | 合约 + 本地部署 + 前端买注通路 |
| 2 | Indexer + DB + 列表 API |
| 3 | 足球 API Sync + 比赛页 |
| 4 | SIWE + 个人中心 + claim |
| 5–6 | 测试网联调、修 bug、演示文档 |

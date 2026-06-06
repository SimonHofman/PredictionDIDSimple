# Prediction DID — 世界杯 Web3 预测市场

基于 DID 的去中心化身份 + 链上预测市场（**Phase 3**：CPMM 流动性、多 outcome、地理围栏、平台统计、对账；Phase 2 Oracle/VC/管理后台保留）。

## 前置要求

- **Node.js** 20 LTS
- **Go** 1.22+
- **Docker Desktop**（PostgreSQL + Redis）
- （可选）本地链：`npx hardhat node`

## 快速开始

### 1. 克隆并配置环境

```powershell
cd PredictionDIDSimple_cursor
copy .env.example .env
copy frontend\.env.example frontend\.env
```

按需编辑 `.env` 与 `frontend/.env`。

### 2. 启动数据库

```powershell
docker compose up -d
```

### 3. 安装依赖

```powershell
cd contracts && npm install && cd ..
cd frontend && npm install && cd ..
cd backend && go mod download && cd ..
```

### 4. 数据库迁移

```powershell
$env:DATABASE_URL="postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable"
cd backend
go run ./cmd/migrate
```

或使用：`.\scripts\dev.ps1 migrate`

### 5. 合约编译与 ABI 同步

```powershell
.\scripts\sync-abi.ps1
```

### 6. 启动 API

```powershell
$env:DATABASE_URL="postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable"
$env:HTTP_PORT="8080"
cd backend
go run ./cmd/api
```

验证：`curl http://localhost:8080/health` → `{"status":"ok"}`

### 7. 启动前端

新终端：

```powershell
cd frontend
npm run dev
```

浏览器打开 http://localhost:5173 ，首页应显示 **API 在线**。

### 8. （可选）本地链 + 部署

终端 A：

```powershell
cd contracts
npx hardhat node
```

终端 B：

```powershell
cd contracts
npx hardhat run scripts/deploy.js --network localhost
```

将输出的 `MarketFactory` 地址记入 `.env` 的 `MARKET_FACTORY_ADDRESS`。

---

## 常用命令

| 命令 | 说明 |
|------|------|
| `make up` | Docker 启动（需 Make，Git Bash/WSL） |
| `.\scripts\dev.ps1 up` | Windows 启动 Docker |
| `.\scripts\dev.ps1 migrate` | 运行迁移 |
| `.\scripts\dev.ps1 api` | 启动 Go API |
| `.\scripts\dev.ps1 web` | 启动前端 |
| `.\scripts\sync-abi.ps1` | 编译合约并同步 ABI |
| `make test-contracts` | 合约测试 |

## 改合约后同步 ABI

```powershell
.\scripts\sync-abi.ps1
```

会更新：

- `backend/pkg/contracts/MarketFactory.json`
- `frontend/src/abis/MarketFactory.json`

## 项目结构

```
contracts/     # Hardhat + Solidity
backend/       # Go API
frontend/      # Vite + React (JavaScript)
docs/          # Phase 规划文档
scripts/       # dev.ps1, sync-abi
```

## Windows 说明

- 推荐使用 **PowerShell** 与 `scripts\dev.ps1`。
- Docker 卷路径使用 WSL2 后端时更稳定；若 Postgres 连接失败，确认端口 `5432` 未被占用。
- 行尾符建议 `core.autocrlf=true`，避免 shell 脚本在 WSL 中执行异常（`sync-abi.sh` 在 Git Bash 下可用）。

## FAQ

**Q: `DATABASE_URL is required`**  
A: 在启动 API 前设置环境变量，或复制根目录 `.env.example` 为 `.env` 并由 IDE 加载。

**Q: `/ready` 返回 503**  
A: Postgres 未启动或迁移未执行。先 `docker compose up -d`，再 `go run ./cmd/migrate`。

**Q: 首页 API 离线**  
A: 确认后端在 `8080` 监听，且 `frontend/.env` 中 `VITE_API_URL=http://localhost:8080`。

**Q: RPC warn / `rpc_ok: false`**  
A: 正常，若未运行 `hardhat node`。启动本地链后约 30 秒内日志会显示 `RPC ok`。

## Phase 3 功能摘要

- **合约**：`PredictionMarketV3`（CPMM + LP）、`MultiOutcomeMarket`（2–8 选项）、`MarketFactoryV3`、`OracleAdapterV2`（m-of-n）
- **后端**：地理围栏、限流、`/stats/platform`、`/markets/:id/pool`、`/kyc/webhook`、`cmd/reconcile`
- **前端**：中英 i18n、合规 Gate、统计页、流动性页、多 outcome 下注
- **文档**：[PRODUCTION-RUNBOOK](./docs/PRODUCTION-RUNBOOK.md)、[DEPLOYMENT-MAINNET](./docs/DEPLOYMENT-MAINNET.md)、[FAQ](./docs/FAQ.md)

部署 V3：`cd contracts && npx hardhat run scripts/deploy-phase3.js --network localhost`

## Phase 2 功能摘要

- **OracleAdapter** + `cmd/oracle`：双源比分校验、冷却、时间锁自动结算
- **VOID** 市场退款、`oracle_jobs` 队列
- **VC** 签发 / 校验、受限市场门禁
- **管理后台** `/admin`、Prometheus `/metrics`、SSE `/events/scores`

详见 [docs/ORACLE-RUNBOOK.md](./docs/ORACLE-RUNBOOK.md)、[docs/ADMIN-GUIDE.md](./docs/ADMIN-GUIDE.md)。

## Phase 1 功能摘要

- **合约**：`MockUSDC`、`MarketFactory`、`PredictionMarket`（Yes/No 互赌池、resolve、claim）
- **后端**：赛程 API、市场 API、SIWE + JWT、`did:pkh` 绑定、链上 Indexer、Mock 世界杯数据同步
- **前端**：wagmi 钱包、下注、持仓、claim、SIWE 登录

详见 [docs/DEPLOYMENT-PHASE1.md](./docs/DEPLOYMENT-PHASE1.md)。

## 文档

- [docs/README.md](./docs/README.md) — Phase 路线图
- [docs/DEPLOYMENT-PHASE1.md](./docs/DEPLOYMENT-PHASE1.md) — Phase 1 部署与演示
- [docs/PHASE-0-TASKS.md](./docs/PHASE-0-TASKS.md) — Phase 0 任务分解
- [docs/PHASE-3-产品与主网.md](./docs/PHASE-3-产品与主网.md) — Phase 3 范围与交付物
- [docs/PRODUCTION-RUNBOOK.md](./docs/PRODUCTION-RUNBOOK.md) — 生产运维

## 技术栈

- 合约：Solidity 0.8.24, Hardhat, OpenZeppelin
- 后端：Go, chi, pgx, golang-migrate, go-ethereum
- 前端：React 18, Vite 5, react-router-dom

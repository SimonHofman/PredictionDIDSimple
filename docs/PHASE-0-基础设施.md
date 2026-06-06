# Phase 0 — 工程基础设施

**周期建议**：约 1 周  
**前置依赖**：无  
**后续 Phase**：[Phase 1 MVP](./PHASE-1-MVP.md)

---

## 1. 目标

搭建可本地一键运行的 Monorepo 骨架，使合约、Go 后端、前端能独立开发与联调，并具备基础 CI。

## 2. 范围

### 2.1 包含

- 目录结构与 README
- Hardhat 初始化（Solidity 0.8.x、OpenZeppelin）
- Go module 初始化（`cmd/api` 占位、`internal` 包结构）
- 前端脚手架（Vite + React + JavaScript）
- `docker-compose`：PostgreSQL、Redis、可选本地链（Hardhat node 或 Anvil）
- 环境变量模板（`.env.example`）
- GitHub Actions：合约 test、Go test、前端 lint/build
- ABI 生成与同步脚本（合约编译 → `backend/pkg/contracts`、`frontend/src/abis`）

### 2.2 不包含

- 业务合约逻辑、真实世界杯 API、生产部署

---

## 3. 任务清单

> **详细分解（20 项、依赖图、验收标准、排期）** → [PHASE-0-TASKS.md](./PHASE-0-TASKS.md)

### 3.1 任务域概览

| 域 | 任务 ID 范围 | 项数 |
|----|----------------|------|
| 仓库根 | P0-ROOT-01 ~ 02 | 2 |
| 智能合约 | P0-CON-01 ~ 04 | 4 |
| 后端 Go | P0-BE-01 ~ 05 | 5 |
| 前端 | P0-FE-01 ~ 04 | 4 |
| DevOps | P0-OPS-01 ~ 05 | 5 |
| 文档与验收 | P0-DOC-01、P0-E2E-01 | 2 |

### 3.2 原任务编号对照

| 原编号 | 新任务 ID |
|--------|-----------|
| C0-1 | P0-CON-01 |
| C0-2 | P0-CON-02 |
| C0-3 | P0-CON-03 |
| C0-4 | P0-CON-04 |
| B0-1 ~ B0-5 | P0-BE-01 ~ 05 |
| F0-1 ~ F0-4 | P0-FE-01 ~ 04 |
| D0-1 ~ D0-4 | P0-OPS-01 ~ 04（另增 OPS-05 ABI 工作流） |

---

## 4. 交付物

- [x] 根目录 README：如何 `docker-compose up`、`hardhat node`、启动 Go 与前端
- [x] 三端项目可本地启动且无编译错误
- [x] CI 配置（`.github/workflows/ci.yml`）
- [x] `docs/` 下 Phase 文档齐全（本文件及后续 Phase）

---

## 5. 验收标准

1. 新开发者按 README 在 30 分钟内跑起 DB + 本地链 + API health + 前端首页。
2. 修改合约并执行脚本后，Go/前端能拿到最新 ABI。
3. 敏感配置仅存在于 `.env`，不进 Git。

---

## 6. 风险与注意事项

- Windows 路径与 Docker 卷映射需文档说明。
- 统一 Node（建议 20 LTS）与 Go（建议 1.22+）版本写在 README。

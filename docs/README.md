# DID Web3 世界杯预测市场 — 文档索引

基于 DID 的去中心化身份 + Web3 预测市场，对接世界杯赛事数据。

## 技术栈

| 层级 | 技术 |
|------|------|
| 智能合约 | Solidity, Hardhat, JavaScript |
| 后端 | Go |
| 前端 | Node.js, JavaScript (React / Vite 或 Next.js) |

## Phase 路线图

| Phase | 名称 | 目标 | 文档 |
|-------|------|------|------|
| 0 | 工程基建 | Monorepo、本地链、CI、基础目录 | [PHASE-0-基础设施.md](./PHASE-0-基础设施.md) · [任务分解](./PHASE-0-TASKS.md) |
| 1 | MVP | 测试网 Yes/No 市场、钱包下注、赛程 API、基础索引 | [PHASE-1-MVP.md](./PHASE-1-MVP.md) |
| 2 | 身份与预言机 | DID/VC、Oracle 结算、双源校验、管理后台 | [PHASE-2-身份与预言机.md](./PHASE-2-身份与预言机.md) · [ORACLE-RUNBOOK](./ORACLE-RUNBOOK.md) |
| 3 | 产品与合规 | 流动性、多市场类型、风控、审计与主网 | [PHASE-3-产品与主网.md](./PHASE-3-产品与主网.md) |

## 全局参考

- [ARCHITECTURE.md](./ARCHITECTURE.md) — 系统架构、模块边界、数据流（跨 Phase 不变部分）

## 依赖关系

```mermaid
flowchart LR
  P0[Phase 0] --> P1[Phase 1]
  P1 --> P2[Phase 2]
  P2 --> P3[Phase 3]
```

每个 Phase 文档均包含：**目标、范围、合约/后端/前端任务清单、交付物、验收标准、风险与排除项**。

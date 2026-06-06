# Oracle 运行手册（Phase 2）

## 组件

| 组件 | 说明 |
|------|------|
| `OracleAdapter` | 链上时间锁 + `resolve` / `void` |
| `cmd/oracle` | Go Worker：双源比分 → 自动结算 |
| `oracle_jobs` 表 | 任务状态机 |

## 环境变量

| 变量 | 说明 |
|------|------|
| `ORACLE_ADAPTER_ADDRESS` | 部署的 OracleAdapter |
| `ORACLE_PRIVATE_KEY` | 持有 `ORACLE_ROLE` 的私钥 |
| `ORACLE_COOLDOWN_MINUTES` | 赛后冷却（建议生产 ≥15） |
| `ORACLE_TIMELOCK_SECONDS` | 链上 confirm 延迟 |
| `MOCK_MATCHES_PATH` | 主数据源 JSON |
| `MOCK_MATCHES_SECONDARY_PATH` | 备数据源 JSON |
| `ALERT_WEBHOOK_URL` | 失败 / manual_review 告警 |

## 启动

```powershell
$env:DATABASE_URL="postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable"
$env:ORACLE_ADAPTER_ADDRESS="0x..."
$env:ORACLE_PRIVATE_KEY="0x..."
cd backend
go run ./cmd/oracle
```

## 状态流

1. `sync` 将比赛标为 `FINISHED`
2. Worker 为关联 `OPEN` 市场创建 `oracle_jobs(pending)`，比赛 → `ORACLE_PENDING`
3. 冷却到期后比对双源比分
4. 一致 → `requestResolve` → 等待 timelock → `confirmResolve` → `confirmed`
5. 不一致 → `manual_review` + Webhook

## 私钥轮换

1. 部署新 Oracle 地址 `grantOracle(new)`
2. 撤销旧地址 `revokeRole`
3. 更新 `.env` 中 `ORACLE_PRIVATE_KEY`
4. 重启 `cmd/oracle`

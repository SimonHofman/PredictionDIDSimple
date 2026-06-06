# 合约升级策略（Phase 2）

本项目 Phase 2 采用 **不可升级代理** 模式：

- 新逻辑 → 部署新 `MarketFactory` / `OracleAdapter`
- 迁移：暂停旧工厂 `createMarket`，新赛事走新地址
- 旧市场继续在原 `PredictionMarket` 实例上结算完毕

若 Phase 3 需要 UUPS：

1. 引入 `OpenZeppelin Upgrades` 插件
2. `MarketFactory` / `OracleAdapter` 改为 UUPS + `timelock` admin
3. 复审计后升级实现合约

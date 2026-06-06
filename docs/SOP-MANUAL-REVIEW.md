# 双源不一致人工处理 SOP

## 触发条件

`oracle_jobs.status = manual_review`，或日志 `ALERT [oracle_manual_review]`。

## 处理步骤

1. 打开 `/admin/oracle` 查看主源 / 备源比分
2. 核对官方赛果（FIFA / 可信 API）
3. 若可确定结果：
   - 修正 `data/mock_matches*.json` 使双源一致
   - 点击 **重试** 或等待 Worker 下一周期
4. 若赛事取消：
   - `POST /admin/markets/:id/void`
   - 比赛状态改为 `CANCELLED`
5. 记录处理人与最终 outcome 到运维日志

## 禁止

- 在未核实赛果前手动调用 `resolve` 脚本
- 将不一致数据强行标为 `confirmed`

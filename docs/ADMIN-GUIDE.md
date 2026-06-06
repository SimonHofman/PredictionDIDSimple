# 管理后台使用说明（内部）

## 访问

1. 设置 `ADMIN_API_KEY` 环境变量
2. 前端 http://localhost:5173/admin/oracle
3. 浏览器控制台：

```javascript
sessionStorage.setItem('admin_key', 'dev-admin-key')
```

## 功能

| 页面 | 路径 | 说明 |
|------|------|------|
| Oracle 队列 | `/admin/oracle` | 查看 pending / manual_review / confirmed |
| 市场配置 | `/admin/markets` | 设置 requires_vc、VOID 市场 |

## API 摘要

- `GET /admin/oracle-jobs`
- `POST /admin/oracle-jobs/:id/retry`
- `POST /admin/markets` — 更新 match 关联市场规则
- `POST /admin/markets/:id/void`
- `POST /credentials/issue`

## 作废市场

对 `CANCELLED` 赛事调用 void，用户可在「我的」页 **退款 Claim**。

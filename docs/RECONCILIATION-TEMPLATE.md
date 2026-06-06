# Weekly Reconciliation Template

| Week | Market address | DB total (yes+no) | Chain USDC balance | Delta | OK |
|------|----------------|-------------------|--------------------|-------|-----|
| YYYY-Www | 0x… | | | | ☐ |

Run: `go run ./cmd/reconcile` from `backend/` with `DATABASE_URL` and `MOCK_USDC_ADDRESS` set.

Investigate any `ok=false` rows in `reconciliation_runs`.

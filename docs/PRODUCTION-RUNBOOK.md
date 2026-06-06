# Production Runbook

## Deploy order

1. Deploy `MockUSDC` or wire production USDC on target chain.
2. Deploy `OracleAdapterV2` (multisig threshold) or keep `OracleAdapter` (timelock) for legacy markets.
3. Deploy `MarketFactoryV3` with collateral + oracle adapter addresses.
4. Export ABIs: `cd contracts && npm run export-abi`.
5. Run DB migrations: API startup or `go run ./cmd/migrate` if present.
6. Start services: `api`, `indexer`, `oracle`, `reconcile` (cron).

## Pause / incident

- **Factory pause**: `MarketFactoryV3.pause()` via owner multisig.
- **Oracle halt**: revoke `ORACLE_ROLE` on adapter; stop `oracle` worker.
- **API degrade**: set `GEO_BLOCK_ENABLED=false` only for internal drills; keep rate limits.

## Rollback

- Frontend: redeploy previous static build from CI artifact.
- Contracts: **not upgradeable** in Phase 3 — deploy new factory and migrate markets manually.
- Database: run `000004_phase3.down.sql` only in non-prod; prod requires DBA review.

## Monitoring

- `/health`, `/ready`, `/metrics`
- Alert on: Oracle job `FAILED`, indexer lag > 50 blocks, reconciliation `ok=false`.

## Secrets

- `JWT_SECRET`, `ADMIN_API_KEY`, `VC_ISSUER_KEY`, `KYC_WEBHOOK_SECRET`, `ORACLE_PRIVATE_KEY` in vault/HSM.
- Never commit `.env` with mainnet keys.

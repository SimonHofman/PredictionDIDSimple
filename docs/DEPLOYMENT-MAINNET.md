# Mainnet / L2 Deployment

## Prerequisites

- Audited contract artifacts (see `AUDIT-CHECKLIST.md`).
- Multisig owner for factory and oracle adapter.
- RPC primary + fallback (`ETH_RPC_URL`, `ETH_RPC_FALLBACK_URL`).

## Steps

1. Set `CHAIN_ID` to target network (e.g. Base `8453`).
2. Fund deployer; run `npx hardhat run scripts/deploy-phase3.js --network <net>`.
3. Verify contracts on block explorer.
4. Set env:
   - `MARKET_FACTORY_V3_ADDRESS`
   - `ORACLE_ADAPTER_V2_ADDRESS`
   - `MOCK_USDC_ADDRESS` (production stablecoin)
5. `INDEXER_START_BLOCK` = factory deploy block.
6. Smoke test: create market, buy, propose resolve with m-of-n, claim.

## Post-deploy

- Publish addresses in `README.md`.
- Schedule `cmd/reconcile` daily cron.
- Enable `GEO_BLOCK_ENABLED=true` and `BLOCKED_COUNTRIES` per legal list.

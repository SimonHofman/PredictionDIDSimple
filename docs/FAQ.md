# FAQ

## How do I claim winnings?

Connect wallet → open market → after status **Resolved**, call **Claim** on-chain. The UI uses the market contract `claim()` function.

## What happens on VOID?

If the oracle voids a market, you can claim your **original stake** (yes + no balances) without winner/loser split.

## CPMM vs parimutuel

- **Phase 1–2** (`PredictionMarket`): parimutuel pools.
- **Phase 3** (`PredictionMarketV3`): constant-product AMM with LP shares and `feeBps`.

## Why is my region blocked?

The API applies geo headers (`CF-IPCountry` / `X-Country-Code`). Restricted countries return HTTP 403. See `/compliance/restricted`.

## KYC

Optional webhook `POST /kyc/webhook` with `X-KYC-Signature` HMAC. On `approved`, a KYC VC can be issued for gated markets.

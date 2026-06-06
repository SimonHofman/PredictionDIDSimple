ALTER TABLE markets ADD COLUMN IF NOT EXISTS market_type TEXT NOT NULL DEFAULT 'BINARY';
ALTER TABLE markets ADD COLUMN IF NOT EXISTS outcome_count SMALLINT NOT NULL DEFAULT 2;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS fee_bps INT NOT NULL DEFAULT 30;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS reserve_yes NUMERIC(78, 0) DEFAULT 0;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS reserve_no NUMERIC(78, 0) DEFAULT 0;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS price_yes_bps INT;

CREATE TABLE IF NOT EXISTS platform_stats_daily (
    day DATE PRIMARY KEY,
    trade_count INT NOT NULL DEFAULT 0,
    trade_volume NUMERIC(78, 0) NOT NULL DEFAULT 0,
    fees_collected NUMERIC(78, 0) NOT NULL DEFAULT 0,
    active_users INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS geo_access_log (
    id BIGSERIAL PRIMARY KEY,
    ip TEXT,
    country_code TEXT,
    allowed BOOLEAN NOT NULL,
    path TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_events (
    id SERIAL PRIMARY KEY,
    external_id TEXT NOT NULL,
    user_address TEXT,
    status TEXT NOT NULL,
    raw_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reconciliation_runs (
    id SERIAL PRIMARY KEY,
    market_address TEXT NOT NULL,
    db_total NUMERIC(78, 0),
    chain_balance NUMERIC(78, 0),
    delta NUMERIC(78, 0),
    ok BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

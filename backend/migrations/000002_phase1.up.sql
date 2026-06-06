CREATE TABLE IF NOT EXISTS matches (
    id              SERIAL PRIMARY KEY,
    external_id     TEXT NOT NULL UNIQUE,
    home_team       TEXT NOT NULL,
    away_team       TEXT NOT NULL,
    kickoff_at      TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'SCHEDULED',
    home_score      INT,
    away_score      INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    address         TEXT NOT NULL UNIQUE,
    did             TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS markets (
    id                  SERIAL PRIMARY KEY,
    match_id            INT REFERENCES matches(id),
    chain_id            BIGINT NOT NULL,
    factory_address     TEXT NOT NULL,
    market_address      TEXT NOT NULL UNIQUE,
    on_chain_market_id  BIGINT NOT NULL,
    match_ref           TEXT NOT NULL,
    question            TEXT NOT NULL,
    end_time            TIMESTAMPTZ NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OPEN',
    winning_outcome     SMALLINT,
    yes_pool            NUMERIC(78, 0) NOT NULL DEFAULT 0,
    no_pool             NUMERIC(78, 0) NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_markets_match_id ON markets(match_id);
CREATE INDEX IF NOT EXISTS idx_markets_status ON markets(status);

CREATE TABLE IF NOT EXISTS trades (
    id              SERIAL PRIMARY KEY,
    market_id       INT NOT NULL REFERENCES markets(id),
    tx_hash         TEXT NOT NULL,
    log_index       INT NOT NULL DEFAULT 0,
    block_number    BIGINT NOT NULL,
    user_address    TEXT NOT NULL,
    outcome         SMALLINT NOT NULL,
    amount          NUMERIC(78, 0) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tx_hash, log_index)
);

CREATE TABLE IF NOT EXISTS positions (
    id              SERIAL PRIMARY KEY,
    market_id       INT NOT NULL REFERENCES markets(id),
    user_address    TEXT NOT NULL,
    yes_amount      NUMERIC(78, 0) NOT NULL DEFAULT 0,
    no_amount       NUMERIC(78, 0) NOT NULL DEFAULT 0,
    claimed         BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (market_id, user_address)
);

CREATE TABLE IF NOT EXISTS indexer_state (
    id              INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    last_block      BIGINT NOT NULL DEFAULT 0,
    factory_address TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO indexer_state (id, last_block) VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;

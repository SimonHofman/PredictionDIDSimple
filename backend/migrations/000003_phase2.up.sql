ALTER TABLE matches ADD COLUMN IF NOT EXISTS resolution_outcome SMALLINT;

ALTER TABLE markets ADD COLUMN IF NOT EXISTS requires_vc BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS restricted_region TEXT;
ALTER TABLE markets ADD COLUMN IF NOT EXISTS resolution_rule TEXT NOT NULL DEFAULT 'HOME_WIN';

CREATE TABLE IF NOT EXISTS oracle_jobs (
    id                  SERIAL PRIMARY KEY,
    match_id            INT REFERENCES matches(id),
    market_id           INT NOT NULL REFERENCES markets(id),
    status              TEXT NOT NULL DEFAULT 'pending',
    primary_home        INT,
    primary_away        INT,
    secondary_home      INT,
    secondary_away      INT,
    proposed_outcome    SMALLINT,
    tx_hash             TEXT,
    error_message       TEXT,
    execute_after       TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oracle_jobs_status ON oracle_jobs(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_oracle_jobs_market_pending
    ON oracle_jobs(market_id) WHERE status IN ('pending', 'submitted', 'manual_review');

CREATE TABLE IF NOT EXISTS credentials (
    id                  SERIAL PRIMARY KEY,
    user_address        TEXT NOT NULL,
    credential_type     TEXT NOT NULL,
    vc_json             JSONB NOT NULL,
    expires_at          TIMESTAMPTZ NOT NULL,
    revoked             BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_credentials_user ON credentials(LOWER(user_address));

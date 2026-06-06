CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS schema_meta (
    id          SERIAL PRIMARY KEY,
    version     TEXT NOT NULL DEFAULT 'phase0',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_meta (version) VALUES ('phase0-init');

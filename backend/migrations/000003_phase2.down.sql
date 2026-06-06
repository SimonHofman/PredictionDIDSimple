DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS oracle_jobs;
ALTER TABLE markets DROP COLUMN IF EXISTS requires_vc;
ALTER TABLE markets DROP COLUMN IF EXISTS restricted_region;
ALTER TABLE markets DROP COLUMN IF EXISTS resolution_rule;
ALTER TABLE matches DROP COLUMN IF EXISTS resolution_outcome;

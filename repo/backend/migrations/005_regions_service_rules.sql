CREATE TABLE IF NOT EXISTS regions (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  parent_region_id BIGINT NULL REFERENCES regions(id) ON DELETE SET NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS service_rules (
  id BIGSERIAL PRIMARY KEY,
  region_id BIGINT NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
  allow_home_pickup BOOLEAN NOT NULL DEFAULT TRUE,
  allow_mail_documents BOOLEAN NOT NULL DEFAULT TRUE,
  blocked BOOLEAN NOT NULL DEFAULT FALSE,
  start_time TIME NULL,
  end_time TIME NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(region_id)
);

CREATE TABLE IF NOT EXISTS blocked_postal_codes (
  id BIGSERIAL PRIMARY KEY,
  service_rule_id BIGINT NOT NULL REFERENCES service_rules(id) ON DELETE CASCADE,
  postal_code TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(service_rule_id, postal_code)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_blocked_postal_code ON blocked_postal_codes(postal_code);

ALTER TABLE addresses
  ADD COLUMN IF NOT EXISTS line1_encrypted TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS line2_encrypted TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS profile_contacts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  relationship TEXT NOT NULL DEFAULT '',
  phone_masked TEXT NOT NULL,
  phone_encrypted TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_profile_contacts_user ON profile_contacts(user_id, created_at DESC);

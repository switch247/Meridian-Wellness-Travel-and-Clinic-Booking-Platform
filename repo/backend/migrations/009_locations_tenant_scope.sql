CREATE TABLE IF NOT EXISTS locations (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO locations(name)
VALUES ('Default Location')
ON CONFLICT (name) DO NOTHING;

ALTER TABLE rooms
  ADD COLUMN IF NOT EXISTS location_id BIGINT REFERENCES locations(id);

UPDATE rooms
SET location_id = (SELECT id FROM locations WHERE name = 'Default Location' LIMIT 1)
WHERE location_id IS NULL;

ALTER TABLE rooms
  ALTER COLUMN location_id SET NOT NULL;

CREATE TABLE IF NOT EXISTS user_locations (
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  location_id BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, location_id)
);

INSERT INTO user_locations(user_id, location_id)
SELECT u.id, l.id
FROM users u
CROSS JOIN locations l
WHERE l.name = 'Default Location'
ON CONFLICT DO NOTHING;

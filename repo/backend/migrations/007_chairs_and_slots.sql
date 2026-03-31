-- Add chairs and chair-aware booking/hold support
BEGIN;

CREATE TABLE IF NOT EXISTS chairs (
  id BIGSERIAL PRIMARY KEY,
  room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE reservation_holds
  ADD COLUMN IF NOT EXISTS chair_id BIGINT REFERENCES chairs(id);

ALTER TABLE bookings
  ADD COLUMN IF NOT EXISTS chair_id BIGINT REFERENCES chairs(id);

CREATE INDEX IF NOT EXISTS idx_reservation_holds_chair_id ON reservation_holds(chair_id);
CREATE INDEX IF NOT EXISTS idx_bookings_chair_id ON bookings(chair_id);

COMMIT;

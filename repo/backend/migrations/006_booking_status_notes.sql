ALTER TABLE bookings ADD COLUMN IF NOT EXISTS session_notes_encrypted TEXT;

ALTER TABLE bookings ALTER COLUMN status SET DEFAULT 'scheduled';
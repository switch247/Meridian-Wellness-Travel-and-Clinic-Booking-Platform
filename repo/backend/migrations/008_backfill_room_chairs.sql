-- Backfill explicit chair records from rooms.chairs_count for existing rooms.
INSERT INTO chairs(room_id, name)
SELECT r.id, CONCAT('Chair ', gs.n)
FROM rooms r
JOIN generate_series(1, 100) AS gs(n) ON gs.n <= GREATEST(COALESCE(r.chairs_count, 0), 0)
LEFT JOIN chairs c ON c.room_id = r.id AND c.name = CONCAT('Chair ', gs.n)
WHERE r.active = TRUE
  AND gs.n <= COALESCE(r.chairs_count, 0)
  AND c.id IS NULL;

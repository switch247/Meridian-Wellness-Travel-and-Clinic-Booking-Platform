INSERT INTO destinations(name, description, image_path, published)
VALUES
  ('Sedona Healing Retreat', 'Wellness-focused desert destination', 'assets/sedona.jpg', TRUE),
  ('Maui Recovery Escape', 'Island recovery and mindful movement', 'assets/maui.jpg', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO packages(destination_id, name, description, published)
SELECT id, name || ' Package', '7-day curated wellness itinerary', TRUE
FROM destinations
ON CONFLICT DO NOTHING;

INSERT INTO package_calendar(package_id, service_date, price_cents, inventory_remaining, blackout_note)
SELECT p.id, CURRENT_DATE + INTERVAL '1 year' + offs.day_offset * INTERVAL '1 day', 129900, 8,
CASE WHEN offs.day_offset % 3 = 0 THEN 'No arrivals after 6:00 PM' ELSE NULL END
FROM packages p
CROSS JOIN (VALUES (1), (2), (3), (4), (5), (6), (7)) AS offs(day_offset)
ON CONFLICT (package_id, service_date) DO NOTHING;

-- Also create some inventory for current dates for testing
INSERT INTO package_calendar(package_id, service_date, price_cents, inventory_remaining, blackout_note)
SELECT p.id, CURRENT_DATE + offs.day_offset, 129900, 8,
CASE WHEN offs.day_offset % 3 = 0 THEN 'No arrivals after 6:00 PM' ELSE NULL END
FROM packages p
CROSS JOIN (VALUES (0), (1), (2), (3), (4), (5), (6), (7)) AS offs(day_offset)
ON CONFLICT (package_id, service_date) DO NOTHING;

INSERT INTO users(username,password_hash,encrypted_phone,encrypted_address)
VALUES (
  'admin',
  '$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO',
  '37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==',
  '4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='
)
ON CONFLICT (username) DO UPDATE SET
  password_hash=EXCLUDED.password_hash,
  encrypted_phone=EXCLUDED.encrypted_phone,
  encrypted_address=EXCLUDED.encrypted_address;

INSERT INTO user_roles(user_id, role_name)
SELECT id, 'admin' FROM users WHERE username='admin'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles(user_id, role_name)
SELECT id, 'operations' FROM users WHERE username='admin'
ON CONFLICT DO NOTHING;

-- Mock placeholder for 3rd-party integrations only.
-- Intentionally mocked because external integrations are not allowed in this offline-first build.

-- Additional development users and roles (idempotent)
INSERT INTO users(username,password_hash,encrypted_phone,encrypted_address)
VALUES
  ('admin@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='),
  ('coach@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='),
  ('clinician@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='),
  ('operations@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='),
  ('traveler1@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc='),
  ('traveler2@example.com','$2a$10$FSRYDfJLqM9x9bPbUc4zKOUOuU2yomTExzw3mfJ3F2qSRgiFeSZIO','37xsZgIybrz6HvxZnXNOeIJJ1xOVpTIomDY0H1eJDDgAwLa/kgsKPA==','4VZMBGmmay/R2tmnS4Mi6sBNGQBTCp7OanVp8nRddsA5hojPeOaggmVQ4CR1OXy8Ng/eypel2k8fdTc=')
ON CONFLICT (username) DO UPDATE SET
  password_hash=EXCLUDED.password_hash,
  encrypted_phone=EXCLUDED.encrypted_phone,
  encrypted_address=EXCLUDED.encrypted_address;

INSERT INTO user_roles(user_id, role_name)
SELECT id, 'admin' FROM users WHERE username='admin@example.com' ON CONFLICT DO NOTHING;
INSERT INTO user_roles(user_id, role_name)
SELECT id, 'coach' FROM users WHERE username='coach@example.com' ON CONFLICT DO NOTHING;
INSERT INTO user_roles(user_id, role_name)
SELECT id, 'clinician' FROM users WHERE username='clinician@example.com' ON CONFLICT DO NOTHING;
INSERT INTO user_roles(user_id, role_name)
SELECT id, 'operations' FROM users WHERE username='operations@example.com' ON CONFLICT DO NOTHING;
INSERT INTO user_roles(user_id, role_name)
SELECT id, 'traveler' FROM users WHERE username='traveler1@example.com' ON CONFLICT DO NOTHING;
INSERT INTO user_roles(user_id, role_name)
SELECT id, 'traveler' FROM users WHERE username='traveler2@example.com' ON CONFLICT DO NOTHING;

INSERT INTO rooms(name, chairs_count, active)
SELECT v.name, v.chairs_count, TRUE
FROM (VALUES ('Room A', 2), ('Room B', 1)) AS v(name, chairs_count)
WHERE NOT EXISTS (SELECT 1 FROM rooms r WHERE r.name=v.name);

INSERT INTO host_availability(host_id, weekday, start_time, end_time, room_id, active)
SELECT u.id, d.wd, '09:00', '17:00', r.id, TRUE
FROM users u
CROSS JOIN (VALUES (1),(2),(3),(4),(5)) d(wd)
JOIN rooms r ON r.name='Room A'
WHERE u.username IN ('coach@example.com', 'clinician@example.com')
ON CONFLICT DO NOTHING;

INSERT INTO host_availability_exceptions(host_id, exception_date, is_available, note)
SELECT u.id, CURRENT_DATE + INTERVAL '1 year' + 1 * INTERVAL '1 day', FALSE, 'Unavailable for seeded exception-day test'
FROM users u
WHERE u.username='coach@example.com'
AND NOT EXISTS (
  SELECT 1 FROM host_availability_exceptions e
  WHERE e.host_id=u.id AND e.exception_date=CURRENT_DATE + INTERVAL '1 year' + 1 * INTERVAL '1 day'
);

-- Also create exception for current date
INSERT INTO host_availability_exceptions(host_id, exception_date, is_available, note)
SELECT u.id, CURRENT_DATE + 1, FALSE, 'Unavailable for seeded exception-day test'
FROM users u
WHERE u.username='coach@example.com'
AND NOT EXISTS (
  SELECT 1 FROM host_availability_exceptions e
  WHERE e.host_id=u.id AND e.exception_date=CURRENT_DATE + 1
);

INSERT INTO holidays(holiday_date, name, closed_all_day)
VALUES (CURRENT_DATE + INTERVAL '1 year' + 30 * INTERVAL '1 day', 'Demo Holiday', TRUE)
ON CONFLICT DO NOTHING;

-- Also create holiday for current period
INSERT INTO holidays(holiday_date, name, closed_all_day)
VALUES (CURRENT_DATE + 30, 'Demo Holiday', TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO routes(destination_id, name, rich_description, image_paths, published)
SELECT d.id, d.name || ' Route', 'Guided wellness route with timed checkpoints.', ARRAY['assets/route1.jpg'], TRUE
FROM destinations d
WHERE NOT EXISTS (SELECT 1 FROM routes r WHERE r.destination_id=d.id AND r.name=d.name || ' Route');

INSERT INTO hotels(destination_id, name, rich_description, image_paths, published)
SELECT d.id, d.name || ' Partner Hotel', 'Partner lodging with wellness amenities.', ARRAY['assets/hotel1.jpg'], TRUE
FROM destinations d
WHERE NOT EXISTS (SELECT 1 FROM hotels h WHERE h.destination_id=d.id AND h.name=d.name || ' Partner Hotel');

INSERT INTO attractions(destination_id, name, rich_description, image_paths, published)
SELECT d.id, d.name || ' Attraction', 'Mindfulness attraction and recovery spot.', ARRAY['assets/attraction1.jpg'], TRUE
FROM destinations d
WHERE NOT EXISTS (SELECT 1 FROM attractions a WHERE a.destination_id=d.id AND a.name=d.name || ' Attraction');

INSERT INTO community_posts(author_user_id, title, body, destination_id, status)
SELECT u.id, 'What should I pack for this retreat?', 'Looking for tips from recent travelers.', d.id, 'active'
FROM users u, destinations d
WHERE u.username='traveler1@example.com'
AND NOT EXISTS (
  SELECT 1 FROM community_posts p
  WHERE p.author_user_id=u.id AND p.title='What should I pack for this retreat?'
)
LIMIT 1
;

INSERT INTO email_template_queue(template_key, recipient_label, subject, body, status)
SELECT 'booking_confirmation', 'traveler1@example.com', 'Your reservation hold', 'Please manually send this confirmation in offline mode.', 'queued'
WHERE NOT EXISTS (
  SELECT 1 FROM email_template_queue WHERE template_key='booking_confirmation' AND recipient_label='traveler1@example.com'
);

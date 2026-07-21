INSERT INTO note (account_id, title, body, created_at, updated_at, deleted_at) VALUES
(101,
 'get-visible-note',
 'Fixture used by the GET note test.',
 '2025-08-12 08:00:00+00',
 '2025-08-12 08:15:00+00',
 NULL),

(101,
 'another note',
 'for list',
 '2025-08-12 09:30:00+00',
 '2025-08-12 09:45:00+00',
 NULL),

(202,
  'update-target',
 'Fixture used by the PUT test.',
 '2025-08-12 10:15:00+00',
 '2025-08-12 10:15:00+00',
 NULL),

(303,
  'soft-delete-target',
 'Fixture used by the DELETE test.',
 '2025-08-12 12:00:00+00',
 '2025-08-12 12:15:00+00',
 NULL);
INSERT INTO app_user(login, password, email,deleted_at) VALUES (
    'alesha',
    '$2a$12$Fkl1bLa5QUNTIud7WV3u6eeQTWnCnutpy6UIvZrHa4gK3a.ph6wOq', --1111
    'alesha@gmail.com',
    NULL
),
(
    'oleg',
    '$2a$12$Fkl1bLa5QUNTIud7WV3u6eeQTWnCnutpy6UIvZrHa4gK3a.ph6wOq', --1111
    'oleg@gmail.com',
    '2026-07-27 19:30:00'
);

INSERT INTO refresh_token (selector,private_hash,user_id,expired_at,revoked) VALUES(
    '18f11757-c9cf-47f4-a187-ddfda409abb4', --normal
    decode('64cee5a6d728d74e087033a11cc86106886e9ee4872c1aa2b5b80443f786f061','hex'), --JZO-pIBrPciCvkUdarnkUWzitwphl3rU5P0xFDPHEo4
    1,
    (now() at time zone 'utc') + interval '30 days',
    false
);


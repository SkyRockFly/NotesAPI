INSERT INTO app_user(login, password, email) VALUES (
    'alesha',
    '$2a$12$Fkl1bLa5QUNTIud7WV3u6eeQTWnCnutpy6UIvZrHa4gK3a.ph6wOq', --1111
    'alesha@gmail.com'
);

INSERT INTO refresh_token (selector,private_hash,user_id,expired_at,revoked) VALUES(
    '18f11757-c9cf-47f4-a187-ddfda409abb4', --normal
    decode('64cee5a6d728d74e087033a11cc86106886e9ee4872c1aa2b5b80443f786f061','hex'), --JZO-pIBrPciCvkUdarnkUWzitwphl3rU5P0xFDPHEo4
    1,
    (now() at time zone 'utc') + interval '30 days',
    false
),
(
    '8c0e3c78-c172-4b5c-b2fe-cc0cd2d795a4', --revoked
    decode('eff6047aa3316d8c2cd931478a1e886166b866482f60fbcee88a462064e0f471','hex'), --lKTh7FUoTpNSkWm477scWKwhnf4CGRosWNX3YcjEH-o
    1,
    (now() at time zone 'utc') + interval '30 days',
    true
),
(
    '00f30850-3ceb-4daf-bd84-33d67e56b775', --expired
    decode('7d6fe9abe4adcbafd658afbfa572e27c074529778f05513e091e01e22dbe88a2','hex'), --SdbVTVauiZLbe5JKDbiIMF-qeDvtD4b6F55rj0aMKrk
    1,
    (now() at time zone 'utc') - interval '2 days',
    false
),
(
    '0049dd76-af66-490d-ac5a-d1d2ac8dfb0d', --rotate and check expire
    decode('f3c966080bbf84c52ac246d45d308993053a04240bec5fff660b5acd9182f946','hex'), --mZYBwwU-fpi-r9L1zPkXf69A-yz2ar1yy-6Pwilhksw
    1,
    (now() at time zone 'utc') + interval '19 days',
    false
);


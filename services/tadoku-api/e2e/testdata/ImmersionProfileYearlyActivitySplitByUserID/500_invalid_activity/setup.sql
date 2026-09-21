insert into logs (id,user_id,language_code,log_activity_id,amount,modifier,computed_score,duration_seconds,eligible_official_leaderboard,created_at,updated_at,deleted_at) values
('aaaaaaaa-aaaa-4aaa-8aaa-000000000001', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 10, 2, null, 60, false, '2026-01-01 00:00:00', '2026-01-01 00:00:00', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000002', '11111111-1111-4111-8111-111111111111', 'eng', 2, 100, 1, 0, 60, false, '2026-01-01 23:59:59', '2026-01-01 23:59:59', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000003', '11111111-1111-4111-8111-111111111111', 'eng', 2, 99, 1, 7.5, 60, false, '2026-12-31 23:59:59', '2026-12-31 23:59:59', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000004', '11111111-1111-4111-8111-111111111111', 'jpn', 3, 3, 1, null, 60, false, '2026-06-01', '2026-06-01', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000005', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 100, 1, null, 60, false, '2025-12-31 23:59:59', '2025-12-31 23:59:59', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000006', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 200, 1, null, 60, false, '2027-01-01', '2027-01-01', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000007', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 999, 1, null, 60, false, '2026-05-01', '2026-05-01', '2026-05-02'),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000008', '11111111-1111-4111-8111-111111111111', 'fra', 4, 4, 1, null, 60, false, '2026-02-01', '2026-02-01', null),
('aaaaaaaa-aaaa-4aaa-8aaa-000000000009', '11111111-1111-4111-8111-111111111111', 'deu', 5, 2, 1, null, 60, false, '2026-02-02', '2026-02-02', null);
insert into logs (id,user_id,language_code,log_activity_id,amount,modifier,duration_seconds,eligible_official_leaderboard,created_at,updated_at) values ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb','44444444-4444-4444-8444-444444444444','jpn',1,555,1,60,false,'2026-01-01','2026-01-01');
update logs set log_activity_id=99;

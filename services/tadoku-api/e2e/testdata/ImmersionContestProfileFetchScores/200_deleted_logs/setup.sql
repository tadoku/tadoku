insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-09-13'),
  ('44444444-4444-4444-8444-444444444444', 'Reader Two', '2026-01-01', '2026-01-01'),
  ('99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01');
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at) values
  ('f1111111-1111-4111-8111-111111111111', '99999999-9999-4999-8999-999999999999', 'stale', false, '2025-12-01', '2026-09-30', '2026-08-31', 'Detach fixture', '{jpn,eng}', '{1}', true, '2026-01-01', '2026-01-01');
insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at) values
  ('e1111111-1111-4111-8111-111111111111', 'f1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', '{jpn,eng}', '2026-01-01', '2026-01-01');
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
insert into contest_logs (contest_id,log_id,amount,modifier,duration_seconds,computed_score) values
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000001', 1, 2, 60, null),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000002', 2, 2, 60, 0),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000003', 3, 2, 60, 11.5),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000004', 4, 2, 60, 3),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000005', 5, 2, 60, null),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000006', 6, 2, 60, null),
('f1111111-1111-4111-8111-111111111111', 'aaaaaaaa-aaaa-4aaa-8aaa-000000000007', 7, 2, 60, 999);
insert into contest_logs (contest_id,log_id,duration_seconds,computed_score) values ('f1111111-1111-4111-8111-111111111111','bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',60,1000);
update logs set deleted_at='2026-09-12';
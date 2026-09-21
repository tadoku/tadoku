insert into logs (id, user_id, language_code, log_activity_id, duration_seconds, eligible_official_leaderboard, created_at, updated_at, deleted_at) values
 ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'jpn', 2, 60, false, '2026-09-01', '2026-09-01', null),
 ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'eng', 2, 60, false, '2026-09-01', '2026-09-01', null),
 ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '11111111-1111-4111-8111-111111111111', 'jpn', 2, 60, false, '2026-09-01', '2026-09-01', null),
 ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', '11111111-1111-4111-8111-111111111111', 'fra', 2, 60, false, '2026-09-01', '2026-09-01', '2026-09-02'),
 ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', '44444444-4444-4444-8444-444444444444', 'deu', 2, 60, false, '2026-09-01', '2026-09-01', null);

insert into log_tags (log_id, user_id, tag)
select 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'custom-' || lpad(n::text, 2, '0')
from generate_series(1, 29) as n;

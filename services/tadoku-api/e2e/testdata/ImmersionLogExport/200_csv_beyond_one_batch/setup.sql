insert into logs (tenant, id, user_id, language_code, log_activity_id, duration_seconds, computed_score, eligible_official_leaderboard, created_at, updated_at)
select 'tadoku/prod', ('c0000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  '11111111-1111-4111-8111-111111111111', 'eng', 2, 60, 1, false, '2026-09-01T00:00:00.123456', '2026-09-01'
from generate_series(1, 501) as n;

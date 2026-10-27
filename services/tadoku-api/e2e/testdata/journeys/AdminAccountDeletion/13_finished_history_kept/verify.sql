select
  contest_logs.contest_id,
  contest_logs.log_id,
  logs.user_id,
  logs.language_code,
  contest_logs.duration_seconds,
  contest_logs.computed_score as score,
  logs.eligible_official_leaderboard as eligible_official
from contest_logs
inner join contests on contests.tenant = contest_logs.tenant and contests.id = contest_logs.contest_id
inner join logs on logs.tenant = contest_logs.tenant and logs.id = contest_logs.log_id
where contest_logs.tenant = 'tadoku/prod'
  and contests.contest_end < '2026-09-12'
  and logs.deleted_at is null
order by contest_logs.contest_id, contest_logs.log_id

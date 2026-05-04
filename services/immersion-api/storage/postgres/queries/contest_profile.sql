-- name: FetchScoresForContestProfile :many
select
  logs.language_code,
  sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as score
from contest_logs
inner join logs
  on logs.id = contest_logs.log_id
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
group by logs.language_code
order by score desc;

-- name: ActivityPerLanguageForContestProfile :many
with eligible_logs as (
  select
    logs.created_at::date as "date",
    logs.language_code,
    coalesce(contest_logs.computed_score, contest_logs.score) as score
  from contest_logs
  inner join logs
    on logs.id = contest_logs.log_id
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.user_id = sqlc.arg('user_id')
    and logs.deleted_at is null
)
select
  "date",
  language_code,
  sum(eligible_logs.score)::real as score
from eligible_logs
group by language_code, "date"
order by "date" asc;
-- name: FetchScoresForContestProfile :many
select
  language_code,
  sum(score)::real as score
from logs
inner join contest_logs
  on contest_logs.log_id = logs.id
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
group by language_code
order by score desc;

-- name: ReadingActivityPerLanguageForContestProfile :many
with eligible_logs as (
  select
    created_at::date as "date",
    language_code,
    score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
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
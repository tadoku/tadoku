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
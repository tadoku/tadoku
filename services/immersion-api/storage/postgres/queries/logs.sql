-- name: CreateLog :one
insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
  tags,
  amount,
  modifier,
  eligible_official_leaderboard,
  "description"
) values (
  sqlc.arg('id'),
  sqlc.arg('user_id'),
  sqlc.arg('language_code'),
  sqlc.arg('log_activity_id'),
  sqlc.arg('unit_id'),
  sqlc.arg('tags'),
  sqlc.arg('amount'),
  sqlc.arg('modifier'),
  sqlc.arg('eligible_official_leaderboard'),
  sqlc.arg('description')
) returning id;

-- name: CreateContestLogRelation :exec
insert into contest_logs (
  contest_id,
  log_id
) values (
  sqlc.arg('contest_id'),
  sqlc.arg('log_id')
);
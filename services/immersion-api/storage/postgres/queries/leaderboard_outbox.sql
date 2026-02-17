-- name: InsertLeaderboardOutboxEvent :exec
insert into leaderboard_outbox (event_type, user_id, contest_id, year)
values (sqlc.arg('event_type'), sqlc.arg('user_id'), sqlc.arg('contest_id'), sqlc.arg('year'));

-- name: FetchAndLockOutboxEvents :many
select id, event_type, user_id, contest_id, year
from leaderboard_outbox
where processed_at is null
order by id
for update skip locked
limit sqlc.arg('batch_size');

-- name: MarkOutboxEventsProcessed :exec
update leaderboard_outbox
set processed_at = now()
where id = any(sqlc.arg('ids')::bigint[]);

-- name: CleanupProcessedOutboxEvents :exec
delete from leaderboard_outbox
where processed_at is not null
  and processed_at < sqlc.arg('before');

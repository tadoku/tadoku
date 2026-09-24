-- name: FetchAndLockLeaderboardOutbox :many
select id, event_type, contest_id, year
from leaderboard_outbox
where processed_at is null
order by id
for update skip locked
limit sqlc.arg('batch_size');

-- name: MarkLeaderboardOutboxProcessed :exec
update leaderboard_outbox
set processed_at = sqlc.arg('processed_at')::timestamp
where id = any(sqlc.arg('ids')::bigint[]);

-- name: CleanupLeaderboardOutbox :exec
delete from leaderboard_outbox
where processed_at is not null and processed_at < sqlc.arg('before')::timestamp;

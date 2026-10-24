-- name: LockContestEventPublication :many
select id
from contests
where id = any(sqlc.arg(contest_ids)::uuid[])
order by id
for no key update;

-- name: InsertContestEvent :exec
insert into contest_events (id, contest_id, log_id, event)
values (sqlc.arg(id), sqlc.arg(contest_id), sqlc.arg(log_id), sqlc.arg(event));

-- name: ReadContestEvents :many
with cursor_event as materialized (
  select events.id, events.position
  from contest_events as events
  where events.contest_id = sqlc.arg(contest_id)
    and events.id = sqlc.narg(cursor_id)::uuid
), checkpoint as (
  select
    case
      when sqlc.arg(from_now)::boolean then (
        select events.id
        from contest_events as events
        where events.contest_id = sqlc.arg(contest_id)
        order by events.position desc
        limit 1
      )
      else sqlc.narg(cursor_id)::uuid
    end as checkpoint_id,
    coalesce((select position from cursor_event), 0)::bigint as position,
    (sqlc.narg(cursor_id)::uuid is null or exists (select 1 from cursor_event))::boolean as cursor_valid
), page as (
  select events.id, events.log_id, events.event, events.position
  from contest_events as events
  cross join checkpoint
  where events.contest_id = sqlc.arg(contest_id)
    and events.position > checkpoint.position
    and checkpoint.cursor_valid
    and not sqlc.arg(from_now)::boolean
  order by events.position
  limit sqlc.arg(page_size)
)
select checkpoint.cursor_valid, checkpoint.checkpoint_id, page.id, page.log_id, page.event
from checkpoint
left join page on true
order by page.position;

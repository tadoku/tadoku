-- name: InsertContestEvents :exec
insert into contest_events (id, contest_id, log_id, event, occurred_at)
select
  unnest(sqlc.arg(ids)::uuid[]),
  unnest(sqlc.arg(contest_ids)::uuid[]),
  unnest(sqlc.arg(log_ids)::uuid[]),
  unnest(sqlc.arg(events)::text[]),
  sqlc.arg(occurred_at);

-- name: ReadContestEvents :many
with cursor_event as materialized (
  select events.id, events.occurred_at
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
        order by events.occurred_at desc, events.id desc
        limit 1
      )
      else sqlc.narg(cursor_id)::uuid
    end as checkpoint_id,
    (select occurred_at from cursor_event) as occurred_at,
    (sqlc.narg(cursor_id)::uuid is null or exists (select 1 from cursor_event))::boolean as cursor_valid
), page as (
  select events.id, events.log_id, events.event, events.occurred_at
  from contest_events as events
  cross join checkpoint
  where events.contest_id = sqlc.arg(contest_id)
    and (
      sqlc.narg(cursor_id)::uuid is null
      or (events.occurred_at, events.id) > (checkpoint.occurred_at, checkpoint.checkpoint_id)
    )
    and checkpoint.cursor_valid
    and not sqlc.arg(from_now)::boolean
  order by events.occurred_at, events.id
  limit sqlc.arg(page_size)
)
select checkpoint.cursor_valid, checkpoint.checkpoint_id, page.id, page.log_id, page.event, page.occurred_at
from checkpoint
left join page on true
order by page.occurred_at, page.id;

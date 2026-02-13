insert into log_tags (log_id, user_id, tag, created_at)
select logs.id, logs.user_id, lower(trim(unnest(logs.tags))), logs.created_at
from logs
where logs.deleted_at is null
  and array_length(logs.tags, 1) > 0
on conflict do nothing;

-- Deduplicate by name (keep first occurrence)
delete from log_default_tags a
using log_default_tags b
where a.ctid > b.ctid and lower(a.name) = lower(b.name);

-- Drop the activity column and index
drop index if exists log_tags_log_activity_id;
alter table log_default_tags drop column log_activity_id;

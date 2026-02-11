begin;

create table moderation_audit_log (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid not null,
  action varchar(100) not null,
  metadata jsonb not null default '{}'::jsonb,
  description text,
  created_at timestamp not null default now()
);

create index moderation_audit_log_action on moderation_audit_log(action);
create index moderation_audit_log_user_id on moderation_audit_log(user_id);

commit;


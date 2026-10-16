begin;

alter table moderation_audit_log
  add column outcome text not null default 'applied'
  constraint moderation_audit_log_outcome check (outcome in ('pending', 'applied', 'failed'));

commit;

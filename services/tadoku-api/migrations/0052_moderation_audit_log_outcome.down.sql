begin;

alter table moderation_audit_log drop column outcome;

commit;

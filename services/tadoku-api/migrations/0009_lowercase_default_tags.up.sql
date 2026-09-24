begin;

update log_default_tags set name = lower(name);

commit;

\if :{?tenant}
\else
  do $$ begin raise exception 'tenant psql variable is required'; end $$;
\endif

begin;
select set_config('tadoku.tenant', :'tenant', true);

insert into profiles (tenant, user_id, created_at, updated_at)
values
  (:'tenant', :'admin_user_id'::uuid, now(), now()),
  (:'tenant', :'reader_user_id'::uuid, now(), now())
on conflict (tenant, user_id) do update
set updated_at = now();

commit;

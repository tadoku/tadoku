begin;

insert into profiles (user_id, created_at, updated_at)
values
  (:'admin_user_id'::uuid, now(), now()),
  (:'reader_user_id'::uuid, now(), now())
on conflict (user_id) do update
set updated_at = now();

commit;

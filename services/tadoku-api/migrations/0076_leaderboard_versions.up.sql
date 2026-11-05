-- Keeps a version per cached leaderboard that every write to its inputs raises
-- in the same transaction, whichever code path or operator issues the write.
-- A board's version is 0 until its first change. The triggers are deferred, so
-- a transaction raises its versions when it commits, after its last write: it
-- takes a transaction-scoped advisory lock for the tenant, then the version
-- rows, so raising cannot deadlock with the transaction's own row locks or with
-- another raise. A write that commits while its test tenant's deletion is in
-- progress fails with lock_not_available instead of waiting on the deletion,
-- so its data and its versions roll back together.
begin;

set local lock_timeout = '5s';

select set_config('search_path', format('%I, pg_temp', current_schema()), true);

create table leaderboard_versions (
  tenant text not null default current_setting('tadoku.tenant'),
  board text not null,
  version bigint not null,

  primary key (tenant, board),
  constraint leaderboard_versions_board_valid check (
    board ~ '^(global|yearly:[0-9]{1,4}|contest:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$'
  ),
  constraint leaderboard_versions_version_positive check (version > 0)
);

alter table leaderboard_versions
  add constraint leaderboard_versions_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table leaderboard_versions validate constraint leaderboard_versions_tenant_fkey;

alter table leaderboard_versions enable row level security;

create policy tenant_isolation on leaderboard_versions
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

create function raise_leaderboard_versions(changed_tenant text, changed_boards text[])
  returns void
  language plpgsql
  set search_path from current
as
$$
begin
  if cardinality(changed_boards) = 0 then
    return;
  end if;
  perform 1 from tenants where tenants.key = changed_tenant for key share nowait;
  if not found then
    return;
  end if;

  perform pg_advisory_xact_lock(hashtextextended('tadoku-leaderboard-versions:' || changed_tenant, 0));

  insert into leaderboard_versions as current (tenant, board, version)
  select changed_tenant, changed.board, 1
  from unnest(changed_boards) as changed (board)
  group by changed.board
  order by changed.board
  on conflict (tenant, board) do update set version = current.version + 1;
end;
$$;

create function leaderboard_versions_from_logs()
  returns trigger
  language plpgsql
  security definer
  set search_path from current
as
$$
declare
  row_tenant text;
  log_ids uuid[];
  old_official boolean := false;
  new_official boolean := false;
  contest_changed boolean := true;
  boards text[] := '{}';
begin
  if tg_op = 'DELETE' then
    row_tenant := old.tenant;
    log_ids := array[old.id];
  else
    row_tenant := new.tenant;
    log_ids := array[new.id];
  end if;
  if tg_op <> 'INSERT' then
    old_official := old.eligible_official_leaderboard and old.deleted_at is null;
  end if;
  if tg_op <> 'DELETE' then
    new_official := new.eligible_official_leaderboard and new.deleted_at is null;
  end if;

  if tg_op = 'UPDATE' then
    log_ids := array[old.id, new.id];
    contest_changed := (old.id, old.user_id, old.deleted_at) is distinct from (new.id, new.user_id, new.deleted_at);
    if old_official and new_official
      and (old.id, old.user_id, coalesce(old.computed_score, old.score))
        is not distinct from (new.id, new.user_id, coalesce(new.computed_score, new.score)) then
      if old.year <> new.year then
        boards := array['yearly:' || old.year, 'yearly:' || new.year];
      end if;
      old_official := false;
      new_official := false;
    end if;
  end if;

  if old_official or new_official then
    boards := boards || 'global'::text;
  end if;
  if old_official then
    boards := boards || ('yearly:' || old.year);
  end if;
  if new_official then
    boards := boards || ('yearly:' || new.year);
  end if;
  if contest_changed then
    boards := boards || array(
      select 'contest:' || contest_logs.contest_id
      from contest_logs
      where contest_logs.tenant = row_tenant and contest_logs.log_id = any(log_ids)
    );
  end if;

  perform raise_leaderboard_versions(row_tenant, boards);
  return null;
end;
$$;

create function leaderboard_versions_from_contest_logs()
  returns trigger
  language plpgsql
  security definer
  set search_path from current
as
$$
begin
  if tg_op = 'INSERT' then
    perform raise_leaderboard_versions(new.tenant, array['contest:' || new.contest_id]);
  elsif tg_op = 'DELETE' then
    perform raise_leaderboard_versions(old.tenant, array['contest:' || old.contest_id]);
  elsif (old.contest_id, old.log_id, coalesce(old.computed_score, old.score))
    is distinct from (new.contest_id, new.log_id, coalesce(new.computed_score, new.score)) then
    perform raise_leaderboard_versions(new.tenant, array['contest:' || old.contest_id, 'contest:' || new.contest_id]);
  end if;
  return null;
end;
$$;

create function leaderboard_versions_from_contest_registrations()
  returns trigger
  language plpgsql
  security definer
  set search_path from current
as
$$
begin
  if tg_op = 'INSERT' then
    if new.deleted_at is null then
      perform raise_leaderboard_versions(new.tenant, array['contest:' || new.contest_id]);
    end if;
  elsif tg_op = 'DELETE' then
    if old.deleted_at is null then
      perform raise_leaderboard_versions(old.tenant, array['contest:' || old.contest_id]);
    end if;
  elsif (old.deleted_at is null or new.deleted_at is null)
    and (old.contest_id, old.user_id, old.deleted_at) is distinct from (new.contest_id, new.user_id, new.deleted_at) then
    perform raise_leaderboard_versions(new.tenant, array['contest:' || old.contest_id, 'contest:' || new.contest_id]);
  end if;
  return null;
end;
$$;

create function leaderboard_versions_from_users()
  returns trigger
  language plpgsql
  security definer
  set search_path from current
as
$$
declare
  row_tenant text;
  user_ids uuid[];
  official_changed boolean := true;
  contest_changed boolean := true;
  boards text[] := '{}';
begin
  if tg_op = 'DELETE' then
    row_tenant := old.tenant;
    user_ids := array[old.id];
  elsif tg_op = 'INSERT' then
    row_tenant := new.tenant;
    user_ids := array[new.id];
  else
    row_tenant := new.tenant;
    user_ids := array[old.id, new.id];
    contest_changed := old.id <> new.id;
    official_changed := contest_changed or old.deleted_at is distinct from new.deleted_at;
  end if;

  if official_changed then
    boards := array(
      select distinct affected.board
      from logs
      cross join lateral (values ('global'), ('yearly:' || logs.year)) as affected (board)
      where logs.tenant = row_tenant
        and logs.user_id = any(user_ids)
        and logs.eligible_official_leaderboard
        and logs.deleted_at is null
    );
  end if;
  if contest_changed then
    boards := boards || array(
      select 'contest:' || contest_registrations.contest_id
      from contest_registrations
      where contest_registrations.tenant = row_tenant and contest_registrations.user_id = any(user_ids)
      union
      select 'contest:' || contests.id
      from contests
      where contests.tenant = row_tenant and contests.owner_user_id = any(user_ids)
    );
  end if;

  perform raise_leaderboard_versions(row_tenant, boards);
  return null;
end;
$$;

create function leaderboard_versions_from_contests()
  returns trigger
  language plpgsql
  security definer
  set search_path from current
as
$$
begin
  if tg_op = 'INSERT' then
    perform raise_leaderboard_versions(new.tenant, array['contest:' || new.id]);
  elsif tg_op = 'DELETE' then
    perform raise_leaderboard_versions(old.tenant, array['contest:' || old.id]);
  elsif (old.id, old.owner_user_id, old.deleted_at) is distinct from (new.id, new.owner_user_id, new.deleted_at) then
    perform raise_leaderboard_versions(new.tenant, array['contest:' || old.id, 'contest:' || new.id]);
  end if;
  return null;
end;
$$;

revoke execute on function raise_leaderboard_versions(text, text[]) from public;
revoke execute on function leaderboard_versions_from_logs() from public;
revoke execute on function leaderboard_versions_from_contest_logs() from public;
revoke execute on function leaderboard_versions_from_contest_registrations() from public;
revoke execute on function leaderboard_versions_from_users() from public;
revoke execute on function leaderboard_versions_from_contests() from public;

create constraint trigger leaderboard_versions_logs
  after insert or update or delete on logs
  deferrable initially deferred
  for each row execute function leaderboard_versions_from_logs();

create constraint trigger leaderboard_versions_contest_logs
  after insert or update or delete on contest_logs
  deferrable initially deferred
  for each row execute function leaderboard_versions_from_contest_logs();

create constraint trigger leaderboard_versions_contest_registrations
  after insert or update or delete on contest_registrations
  deferrable initially deferred
  for each row execute function leaderboard_versions_from_contest_registrations();

create constraint trigger leaderboard_versions_users
  after insert or update or delete on users
  deferrable initially deferred
  for each row execute function leaderboard_versions_from_users();

create constraint trigger leaderboard_versions_contests
  after insert or update or delete on contests
  deferrable initially deferred
  for each row execute function leaderboard_versions_from_contests();

commit;

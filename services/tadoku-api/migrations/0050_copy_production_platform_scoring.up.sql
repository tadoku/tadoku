-- Gives a new test tenant production's active platform scoring rule set.
-- tadoku-tenant provision calls it after inserting the test registry row; the
-- infrastructure grants execute to tadoku_tenant_lifecycle.
begin;

set local lock_timeout = '5s';

select set_config('search_path', format('%I, pg_temp', current_schema()), true);

create function copy_production_platform_scoring(target_tenant text)
  returns void
  language plpgsql
  security definer
  set search_path from current
as
$$
declare
  production_rule_set_id uuid;
  copied_rule_set_id uuid := gen_random_uuid();
begin
  if not exists (
    select
    from tenants
    where key = target_tenant
      and kind = 'test'
  ) then
    raise exception 'tenant % is not a registered test tenant', target_tenant;
  end if;

  if exists (select from platform_scoring_config where tenant = target_tenant) then
    return;
  end if;

  select scoring_rule_sets.id
  into production_rule_set_id
  from platform_scoring_config
  inner join scoring_rule_sets on scoring_rule_sets.id = platform_scoring_config.active_rule_set_id
  where platform_scoring_config.tenant = 'tadoku/prod'
    and scoring_rule_sets.tenant = 'tadoku/prod'
    and scoring_rule_sets.scope = 'platform';

  if production_rule_set_id is null then
    raise exception 'tadoku/prod has no active platform scoring rule set';
  end if;

  insert into scoring_rule_sets (tenant, id, scope, version, status, created_at, published_at)
  select target_tenant, copied_rule_set_id, scope, version, status, created_at, published_at
  from scoring_rule_sets
  where id = production_rule_set_id;

  insert into scoring_rules (
    tenant,
    rule_set_id,
    priority,
    stackable,
    activity_id,
    unit_key,
    language_code,
    tag,
    score_source,
    rate
  )
  select
    target_tenant,
    copied_rule_set_id,
    priority,
    stackable,
    activity_id,
    unit_key,
    language_code,
    tag,
    score_source,
    rate
  from scoring_rules
  where rule_set_id = production_rule_set_id;

  insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
  values (target_tenant, true, copied_rule_set_id);
end;
$$;

revoke execute on function copy_production_platform_scoring(text) from public;

commit;

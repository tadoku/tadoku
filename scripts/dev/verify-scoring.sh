#!/usr/bin/env bash
set -euo pipefail

DB_NAME="tadoku-dev-db"
DB_NAMESPACE="${TADOKU_DEV_NAMESPACE:-default}"
DB_PASSWORD="${TADOKU_DEV_DB_PASSWORD:-dev-foobar}"
API_NAMESPACE="${TADOKU_IMMERSION_NAMESPACE:-tdk-immersion-api}"
ALLOWED_SHADOW_MISMATCHES="${SCORING_ALLOWED_SHADOW_MISMATCHES:-0}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd kubectl

db_pod="$(kubectl -n "$DB_NAMESPACE" get pod \
  -l "application=spilo,cluster-name=${DB_NAME},spilo-role=master" \
  -o jsonpath='{.items[0].metadata.name}')"

if [ -z "$db_pod" ]; then
  echo "could not find the primary ${DB_NAME} pod in namespace ${DB_NAMESPACE}" >&2
  exit 1
fi

database_url="postgres://immersion:${DB_PASSWORD}@${DB_NAME}.${DB_NAMESPACE}/immersion?sslmode=require"

echo "checking active rule coverage, parity rates, and provenance invariants..."
kubectl -n "$DB_NAMESPACE" exec -i "$db_pod" -- env PGPASSWORD="$DB_PASSWORD" \
  psql -X -q "$database_url" -v ON_ERROR_STOP=1 <<'SQL'
do $$
declare
  mismatch_count integer;
  duration_mismatch_count integer;
  invalid_provenance_count integer;
begin
  if not exists (
    select 1
    from platform_scoring_config
    inner join scoring_rule_sets
      on scoring_rule_sets.id = platform_scoring_config.active_rule_set_id
    where platform_scoring_config.singleton = true
      and scoring_rule_sets.scope = 'platform'
      and scoring_rule_sets.status = 'published'
  ) then
    raise exception 'no published active platform scoring rule set';
  end if;

  with active_rules as (
    select scoring_rules.*
    from platform_scoring_config
    inner join scoring_rules
      on scoring_rules.rule_set_id = platform_scoring_config.active_rule_set_id
    where platform_scoring_config.singleton = true
  ),
  unit_rates as (
    select
      log_units.id,
      log_units.log_activity_id,
      log_units.unit_key,
      log_units.modifier as legacy_rate,
      coalesce((
        select base.rate
        from active_rules as base
        where base.stackable = false
          and base.score_source = 'amount'
          and base.activity_id = log_units.log_activity_id
          and (base.unit_key is null or base.unit_key = log_units.unit_key)
          and (
            base.language_code is null
            or base.language_code = log_units.language_code
          )
          and base.tag is null
        order by base.priority
        limit 1
      ), 0) * coalesce((
        select
          case
            when bool_or(modifier.rate = 0) then 0
            else exp(sum(ln(modifier.rate)))
          end
        from active_rules as modifier
        where modifier.stackable = true
          and modifier.score_source = 'amount'
          and modifier.activity_id = log_units.log_activity_id
          and (
            modifier.unit_key is null
            or modifier.unit_key = log_units.unit_key
          )
          and (
            modifier.language_code is null
            or modifier.language_code = log_units.language_code
          )
          and modifier.tag is null
      ), 1) as engine_rate
    from log_units
  )
  select count(*)
  into mismatch_count
  from unit_rates
  where abs(engine_rate - legacy_rate) > 0.00001
    and not (
      log_activity_id = 2
      and (
        (
          unit_key = 'listening_minute'
          and abs(legacy_rate - 0.5) <= 0.00001
          and abs(engine_rate - 0.4) <= 0.00001
        )
        or (
          unit_key = 'listening_dense_minutes'
          and abs(legacy_rate - 0.7) <= 0.00001
          and abs(engine_rate - 0.6) <= 0.00001
        )
      )
    );

  if mismatch_count > 0 then
    raise exception '% legacy unit rates differ from the active scoring rules', mismatch_count;
  end if;

  with expected(activity_id, rate) as (
    values
      (1::smallint, 0.2::real),
      (2::smallint, 0.4::real),
      (3::smallint, 0.2::real),
      (4::smallint, 0.5::real),
      (5::smallint, 0.5::real)
  ),
  active_rules as (
    select scoring_rules.*
    from platform_scoring_config
    inner join scoring_rules
      on scoring_rules.rule_set_id = platform_scoring_config.active_rule_set_id
    where platform_scoring_config.singleton = true
  )
  select count(*)
  into duration_mismatch_count
  from expected
  where not exists (
    select 1
    from active_rules
    where active_rules.activity_id = expected.activity_id
      and active_rules.stackable = false
      and active_rules.unit_key is null
      and active_rules.language_code is null
      and active_rules.tag is null
      and active_rules.score_source = 'duration_minutes'
      and abs(active_rules.rate - expected.rate) <= 0.00001
  );

  if duration_mismatch_count > 0 then
    raise exception '% documented duration rates are missing or incorrect', duration_mismatch_count;
  end if;

  select
    (select count(*) from logs
      where score_rule_ids is not null
        and cardinality(score_rule_ids) <> cardinality(score_rates))
    +
    (select count(*) from contest_logs
      where score_rule_ids is not null
        and cardinality(score_rule_ids) <> cardinality(score_rates))
  into invalid_provenance_count;

  if invalid_provenance_count > 0 then
    raise exception '% score snapshots have invalid provenance arrays', invalid_provenance_count;
  end if;
end;
$$;
SQL

echo "checking recent shadow diagnostics..."
shadow_mismatches="$(
  kubectl -n "$API_NAMESPACE" logs deployment/immersion-api \
    --since=24h 2>/dev/null |
    grep 'scoring shadow mismatch' |
    grep -Evc 'activity_id(=|":)2' || true
)"

if [ "$shadow_mismatches" -gt "$ALLOWED_SHADOW_MISMATCHES" ]; then
  echo "found ${shadow_mismatches} scoring shadow mismatches (allowed: ${ALLOWED_SHADOW_MISMATCHES})" >&2
  exit 1
fi

echo "scoring verification passed (shadow mismatches: ${shadow_mismatches})"

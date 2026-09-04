#!/usr/bin/env bash

set -euo pipefail
umask 077

usage() {
  cat <<'EOF'
Usage: copy-live-data.sh <content|profile|authz> <backup.sql>

Required environment:
  DATABASE_CONVERGENCE_SOURCE_DSN       Source PostgreSQL connection string
  DATABASE_CONVERGENCE_TARGET_DSN       Target PostgreSQL connection string
  DATABASE_CONVERGENCE_WRITES_PAUSED=yes

Optional environment:
  DATABASE_CONVERGENCE_SCHEMA           Source and target schema (default: data)

The backup path must not already exist. The script retains that scoped pg_dump
after copying the selected service's data to the canonical database.
EOF
}

fail() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

source_psql() {
  psql --dbname="$source_dsn" -X -v ON_ERROR_STOP=1 "$@"
}

target_psql() {
  psql --dbname="$target_dsn" -X -v ON_ERROR_STOP=1 "$@"
}

fingerprint() {
  local dsn="$1"
  local table="$2"
  local key="$3"

  psql --dbname="$dsn" -X -A -t -v ON_ERROR_STOP=1 -c \
    "select count(*)::text || ':' || coalesce(md5(string_agg(md5(to_jsonb(t)::text), '' order by t.${key})), md5('')) from ${schema}.${table} as t"
}

append_fingerprint_check() {
  local file="$1"
  local relation="$2"
  local key="$3"
  local expected="$4"
  local label="$5"

  cat >>"$file" <<EOF
do \$validation\$
declare
  actual text;
begin
  select count(*)::text || ':' || coalesce(md5(string_agg(md5(to_jsonb(t)::text), '' order by t.${key})), md5(''))
    into actual
    from ${relation} as t;
  if actual <> '${expected}' then
    raise exception '${label} fingerprint mismatch: expected %, got %', '${expected}', actual;
  end if;
end
\$validation\$;
EOF
}

if [ "$#" -ne 2 ]; then
  usage >&2
  exit 2
fi

service="$1"
backup_path="$2"
source_dsn="${DATABASE_CONVERGENCE_SOURCE_DSN:-}"
target_dsn="${DATABASE_CONVERGENCE_TARGET_DSN:-}"
schema="${DATABASE_CONVERGENCE_SCHEMA:-data}"

[ -n "$source_dsn" ] || fail "DATABASE_CONVERGENCE_SOURCE_DSN is required"
[ -n "$target_dsn" ] || fail "DATABASE_CONVERGENCE_TARGET_DSN is required"
[ "${DATABASE_CONVERGENCE_WRITES_PAUSED:-}" = "yes" ] || \
  fail "pause and drain service writes, then set DATABASE_CONVERGENCE_WRITES_PAUSED=yes"
[[ "$schema" =~ ^[a-z_][a-z0-9_]*$ ]] || fail "DATABASE_CONVERGENCE_SCHEMA is not a safe PostgreSQL identifier"
[ ! -e "$backup_path" ] || fail "backup already exists: $backup_path"
[ -d "$(dirname -- "$backup_path")" ] || fail "backup directory does not exist: $(dirname -- "$backup_path")"

require_command pg_dump
require_command psql
require_command sed
require_command sha256sum

case "$service" in
  content)
    source_version=3
    tables=(pages pages_content posts posts_content announcements)
    keys=(id id id id id)
    ;;
  profile)
    source_version=2
    tables=(profiles account_deletion_requests)
    keys=(user_id id)
    ;;
  authz)
    source_version=2
    tables=(moderation_audit_log)
    keys=(id)
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac

source_database="$(source_psql -A -t -c 'select current_database()')"
target_database="$(target_psql -A -t -c 'select current_database()')"
[ "$source_database" != "$target_database" ] || fail "source and target resolve to the same database"

source_ledger="$(source_psql -A -t -c "select version::text || ':' || dirty::text from ${schema}.schema_migrations")"
target_ledger="$(target_psql -A -t -c "select version::text || ':' || dirty::text from ${schema}.schema_migrations")"
[ "$source_ledger" = "${source_version}:false" ] || \
  fail "unexpected $service source migration state (expected ${source_version}:false, got $source_ledger)"
[ "$target_ledger" = "29:false" ] || \
  fail "unexpected target migration state (expected 29:false, got $target_ledger)"

fingerprints=()
for i in "${!tables[@]}"; do
  value="$(fingerprint "$source_dsn" "${tables[$i]}" "${keys[$i]}")"
  [[ "$value" =~ ^[0-9]+:[0-9a-f]{32}$ ]] || fail "could not fingerprint ${tables[$i]}"
  fingerprints+=("$value")
done

partial_backup="$(mktemp "${backup_path}.partial.XXXXXX")"
import_sql="$(mktemp "${TMPDIR:-/tmp}/tadoku-database-convergence.XXXXXX.sql")"
cleanup() {
  if [ -n "$partial_backup" ]; then
    rm -f -- "$partial_backup"
  fi
  rm -f -- "$import_sql"
}
trap cleanup EXIT

dump_tables=()
for table in "${tables[@]}"; do
  dump_tables+=("--table=${schema}.${table}")
done

pg_dump --dbname="$source_dsn" \
  --data-only \
  --no-owner \
  --no-privileges \
  "${dump_tables[@]}" \
  >"$partial_backup"
mv -- "$partial_backup" "$backup_path"
partial_backup=""

if [ "$service" = "authz" ]; then
  cat >"$import_sql" <<EOF
create temporary table moderation_audit_log_stage
  (like ${schema}.moderation_audit_log including defaults)
  on commit drop;
EOF
  sed "s/^COPY ${schema}\.moderation_audit_log (/COPY moderation_audit_log_stage (/" \
    "$backup_path" >>"$import_sql"
  grep -q '^COPY moderation_audit_log_stage (' "$import_sql" || \
    fail "could not prepare the Authz staging import"
  if grep -q "^COPY ${schema}\.moderation_audit_log (" "$import_sql"; then
    fail "Authz dump still targets the live table"
  fi
  append_fingerprint_check \
    "$import_sql" moderation_audit_log_stage id "${fingerprints[0]}" "authz staged source"
  cat >>"$import_sql" <<EOF
do \$validation\$
begin
  if exists (
    select 1
      from moderation_audit_log_stage as source
      join ${schema}.moderation_audit_log as target using (id)
     where row(source.user_id, source.action, source.metadata, source.description, source.created_at)
       is distinct from
       row(target.user_id, target.action, target.metadata, target.description, target.created_at)
  ) then
    raise exception 'authz audit UUID conflict';
  end if;
end
\$validation\$;

insert into ${schema}.moderation_audit_log
  (id, user_id, action, metadata, description, created_at)
select id, user_id, action, metadata, description, created_at
  from moderation_audit_log_stage
on conflict (id) do nothing;

do \$validation\$
begin
  if exists (
    select 1
      from moderation_audit_log_stage as source
      left join ${schema}.moderation_audit_log as target using (id)
     where target.id is null
        or row(source.user_id, source.action, source.metadata, source.description, source.created_at)
          is distinct from
          row(target.user_id, target.action, target.metadata, target.description, target.created_at)
  ) then
    raise exception 'authz audit validation failed';
  end if;
end
\$validation\$;
EOF
else
  {
    printf 'lock table '
    for i in "${!tables[@]}"; do
      [ "$i" -eq 0 ] || printf ', '
      printf '%s.%s' "$schema" "${tables[$i]}"
    done
    printf ' in access exclusive mode;\n'
    printf 'do $validation$\nbegin\n'
    for table in "${tables[@]}"; do
      printf '  if exists (select 1 from %s.%s) then raise exception '\''target table %s is not empty'\''; end if;\n' \
        "$schema" "$table" "$table"
    done
    printf 'end\n$validation$;\n'
    cat "$backup_path"
  } >"$import_sql"

  for i in "${!tables[@]}"; do
    append_fingerprint_check \
      "$import_sql" "${schema}.${tables[$i]}" "${keys[$i]}" "${fingerprints[$i]}" "${tables[$i]}"
  done
fi

target_psql --single-transaction --file="$import_sql" >/dev/null

echo "copied $service data from $source_database to $target_database"
for i in "${!tables[@]}"; do
  echo "${tables[$i]} rows: ${fingerprints[$i]%%:*}"
done
echo "backup: $backup_path"
echo "backup sha256: $(sha256sum "$backup_path" | cut -d ' ' -f 1)"

#!/usr/bin/env bash

set -euo pipefail
umask 077

usage() {
  cat <<'EOF'
Usage: copy-database.sh <backup.dump>

Required environment:
  ORY_COPY_SOURCE_DSN
  ORY_COPY_TARGET_DSN
  ORY_COPY_WRITES_PAUSED=yes

Optional environment:
  ORY_COPY_SCHEMA                 Ory schema (default: data)

Copies one Kratos or Keto data schema. The target schema must be empty, its
role search_path must include the Ory schema, and the backup path must not
exist. The protected custom-format dump is retained after restore.
EOF
}

fail() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}

query() {
  local dsn="$1"
  local sql="$2"
  psql --dbname="$dsn" -X -A -t -v ON_ERROR_STOP=1 -c "$sql"
}

server_identity() {
  query "$1" "select current_database() || '|' || coalesce(inet_server_addr()::text, 'local') || '|' || coalesce(inet_server_port()::text, 'local')"
}

ledger() {
  query "$1" "select version::text || ':' || version_self::text from ${schema}.schema_migration order by version, version_self"
}

table_names() {
  query "$1" "select tablename from pg_tables where schemaname = '${schema}' order by tablename"
}

schema_inventory() {
  query "$1" "select pg_class.relkind::text || ':' || pg_class.relname from pg_class join pg_namespace on pg_namespace.oid = pg_class.relnamespace where pg_namespace.nspname = '${schema}' and pg_class.relkind in ('r', 'p', 'i', 'S') order by pg_class.relkind, pg_class.relname"
}

sequence_values() {
  query "$1" "select sequencename || ':' || coalesce(last_value::text, '') from pg_sequences where schemaname = '${schema}' order by sequencename"
}

fingerprints() {
  local dsn="$1"
  local table
  while IFS= read -r table; do
    [ -n "$table" ] || continue
    [[ "$table" =~ ^[a-z_][a-z0-9_]*$ ]] || fail "database contains an unsupported table name: $table"
    printf '%s|' "$table"
    query "$dsn" "select count(*)::text || ':' || coalesce(md5(string_agg(md5(to_jsonb(t)::text), '' order by md5(to_jsonb(t)::text))), md5('')) from ${schema}.${table} as t"
  done < <(table_names "$dsn")
}

if [ "$#" -ne 1 ]; then
  usage >&2
  exit 2
fi

backup_path="$1"
source_dsn="${ORY_COPY_SOURCE_DSN:-}"
target_dsn="${ORY_COPY_TARGET_DSN:-}"
schema="${ORY_COPY_SCHEMA:-data}"

[ -n "$source_dsn" ] || fail "ORY_COPY_SOURCE_DSN is required"
[ -n "$target_dsn" ] || fail "ORY_COPY_TARGET_DSN is required"
[ "$source_dsn" != "$target_dsn" ] || fail "source and target DSNs must differ"
[ "${ORY_COPY_WRITES_PAUSED:-}" = "yes" ] || fail "pause and drain writes, then set ORY_COPY_WRITES_PAUSED=yes"
[[ "$schema" =~ ^[a-z_][a-z0-9_]*$ ]] || fail "ORY_COPY_SCHEMA is not a safe PostgreSQL identifier"
[ ! -e "$backup_path" ] || fail "backup already exists: $backup_path"
[ -d "$(dirname -- "$backup_path")" ] || fail "backup directory does not exist: $(dirname -- "$backup_path")"

for command in psql pg_dump pg_restore sha256sum cmp; do
  require_command "$command"
done

source_identity="$(server_identity "$source_dsn")"
target_identity="$(server_identity "$target_dsn")"
[ "$source_identity" != "$target_identity" ] || fail "source and target resolve to the same database"

target_search_path_ok="$(query "$target_dsn" "select exists (select 1 from unnest(string_to_array(current_setting('search_path'), ',')) as item where btrim(item, ' \"') = '${schema}')")"
[ "$target_search_path_ok" = "t" ] || fail "target role search_path does not include $schema"

source_ledger="$(ledger "$source_dsn")"
[ -n "$source_ledger" ] || fail "source Ory migration ledger is empty"

target_relations="$(query "$target_dsn" "select count(*) from pg_class join pg_namespace on pg_namespace.oid = pg_class.relnamespace where pg_namespace.nspname = '${schema}' and pg_class.relkind in ('r', 'p', 'v', 'm', 'S')")"
[ "$target_relations" = "0" ] || fail "target schema $schema is not empty"

partial_backup="$(mktemp "${backup_path}.partial.XXXXXX")"
source_fingerprints="$(mktemp "${TMPDIR:-/tmp}/ory-source-fingerprints.XXXXXX")"
target_fingerprints="$(mktemp "${TMPDIR:-/tmp}/ory-target-fingerprints.XXXXXX")"
cleanup() {
  rm -f -- "$partial_backup" "$source_fingerprints" "$target_fingerprints"
}
trap cleanup EXIT

fingerprints "$source_dsn" >"$source_fingerprints"
source_inventory="$(schema_inventory "$source_dsn")"
source_sequences="$(sequence_values "$source_dsn")"
pg_dump --dbname="$source_dsn" --schema="$schema" --format=custom --no-owner --no-privileges --file="$partial_backup"
mv -- "$partial_backup" "$backup_path"
partial_backup=""

pg_restore --dbname="$target_dsn" --exit-on-error --single-transaction --no-owner --no-privileges "$backup_path"

target_ledger="$(ledger "$target_dsn")"
[ "$target_ledger" = "$source_ledger" ] || fail "Ory migration ledger differs after restore"
[ "$(schema_inventory "$target_dsn")" = "$source_inventory" ] || fail "schema object inventory differs after restore"
[ "$(sequence_values "$target_dsn")" = "$source_sequences" ] || fail "sequence values differ after restore"

fingerprints "$target_dsn" >"$target_fingerprints"
cmp --silent "$source_fingerprints" "$target_fingerprints" || fail "table fingerprints differ after restore"

echo "copied Ory schema $schema from $source_identity to $target_identity"
echo "migration rows: $(printf '%s\n' "$target_ledger" | wc -l | tr -d ' ')"
echo "backup: $backup_path"
echo "backup sha256: $(sha256sum "$backup_path" | cut -d ' ' -f 1)"

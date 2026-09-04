#!/usr/bin/env bash

set -euo pipefail
umask 077

usage() {
  cat <<'EOF'
Usage: copy-database.sh <backup.dump>

Required environment:
  TADOKU_COPY_SOURCE_DSN
  TADOKU_COPY_TARGET_DSN
  TADOKU_COPY_EXPECTED_VERSION
  TADOKU_COPY_WRITES_PAUSED=yes

Optional environment:
  TADOKU_COPY_SCHEMA                 Application schema (default: data)

The target application schema must be empty and the backup path must not exist.
The script retains the protected custom-format dump after a transactional restore.
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
  query "$1" "select version::text || ':' || dirty::text from ${schema}.schema_migrations"
}

table_names() {
  query "$1" "select tablename from pg_tables where schemaname = '${schema}' and tablename <> 'schema_migrations' order by tablename"
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
source_dsn="${TADOKU_COPY_SOURCE_DSN:-}"
target_dsn="${TADOKU_COPY_TARGET_DSN:-}"
expected_version="${TADOKU_COPY_EXPECTED_VERSION:-}"
schema="${TADOKU_COPY_SCHEMA:-data}"

[ -n "$source_dsn" ] || fail "TADOKU_COPY_SOURCE_DSN is required"
[ -n "$target_dsn" ] || fail "TADOKU_COPY_TARGET_DSN is required"
[ "$source_dsn" != "$target_dsn" ] || fail "source and target DSNs must differ"
[[ "$expected_version" =~ ^[1-9][0-9]*$ ]] || fail "TADOKU_COPY_EXPECTED_VERSION must be a positive integer"
[ "${TADOKU_COPY_WRITES_PAUSED:-}" = "yes" ] || fail "pause and drain writes, then set TADOKU_COPY_WRITES_PAUSED=yes"
[[ "$schema" =~ ^[a-z_][a-z0-9_]*$ ]] || fail "TADOKU_COPY_SCHEMA is not a safe PostgreSQL identifier"
[ ! -e "$backup_path" ] || fail "backup already exists: $backup_path"
[ -d "$(dirname -- "$backup_path")" ] || fail "backup directory does not exist: $(dirname -- "$backup_path")"

for command in psql pg_dump pg_restore sha256sum cmp; do
  require_command "$command"
done

source_identity="$(server_identity "$source_dsn")"
target_identity="$(server_identity "$target_dsn")"
[ "$source_identity" != "$target_identity" ] || fail "source and target resolve to the same database"

source_ledger="$(ledger "$source_dsn")"
[ "$source_ledger" = "${expected_version}:false" ] || \
  fail "unexpected source migration state (expected ${expected_version}:false, got $source_ledger)"

target_relations="$(query "$target_dsn" "select count(*) from pg_class join pg_namespace on pg_namespace.oid = pg_class.relnamespace where pg_namespace.nspname = '${schema}' and pg_class.relkind in ('r', 'p', 'v', 'm', 'S')")"
[ "$target_relations" = "0" ] || fail "target schema $schema is not empty"

partial_backup="$(mktemp "${backup_path}.partial.XXXXXX")"
source_fingerprints="$(mktemp "${TMPDIR:-/tmp}/tadoku-source-fingerprints.XXXXXX")"
target_fingerprints="$(mktemp "${TMPDIR:-/tmp}/tadoku-target-fingerprints.XXXXXX")"
cleanup() {
  rm -f -- "$partial_backup" "$source_fingerprints" "$target_fingerprints"
}
trap cleanup EXIT

fingerprints "$source_dsn" >"$source_fingerprints"
pg_dump \
  --dbname="$source_dsn" \
  --format=custom \
  --no-owner \
  --no-privileges \
  --schema="$schema" \
  --extension=uuid-ossp \
  --file="$partial_backup"
mv -- "$partial_backup" "$backup_path"
partial_backup=""

pg_restore --dbname="$target_dsn" --exit-on-error --single-transaction --no-owner --no-privileges "$backup_path"

target_ledger="$(ledger "$target_dsn")"
[ "$target_ledger" = "$source_ledger" ] || \
  fail "target migration state differs after restore (source $source_ledger, target $target_ledger)"

fingerprints "$target_dsn" >"$target_fingerprints"
cmp --silent "$source_fingerprints" "$target_fingerprints" || fail "table fingerprints differ after restore"

echo "copied $source_identity to $target_identity"
echo "migration state: $target_ledger"
echo "backup: $backup_path"
echo "backup sha256: $(sha256sum "$backup_path" | cut -d ' ' -f 1)"

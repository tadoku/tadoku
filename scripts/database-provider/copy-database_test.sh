#!/usr/bin/env bash

set -euo pipefail

copy_script="$1"
if [ ! -x "$copy_script" ]; then
  copy_script="${RUNFILES_DIR:?}/$copy_script"
fi
test_directory="$(mktemp -d)"
trap 'rm -rf -- "$test_directory"' EXIT
fake_bin="$test_directory/bin"
mkdir "$fake_bin"

cat >"$fake_bin/psql" <<'EOF'
#!/usr/bin/env bash
dsn=""
sql=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dbname=*) dsn="${1#--dbname=}" ;;
    -c) shift; sql="$1" ;;
  esac
  shift
done

case "$sql" in
  *current_database*)
    if [ "${FAKE_SAME_IDENTITY:-}" = yes ]; then echo 'same|10.0.0.1|5432'
    elif [ "$dsn" = source ]; then echo 'source|10.0.0.1|5432'
    else echo 'target|10.0.0.2|5432'
    fi
    ;;
  *current_setting*)
    if [ "${FAKE_BAD_SEARCH_PATH:-}" = yes ]; then echo f; else echo t; fi
    ;;
  *'select count(*) from pg_class'*) echo "${FAKE_TARGET_RELATIONS:-0}" ;;
  *'pg_class.relkind::text'*)
    if [ "$dsn" = target ] && [ "${FAKE_TARGET_INVENTORY_DIFFERENT:-}" = yes ]; then
      printf 'r:schema_migration\nr:sessions\n'
    else
      printf 'i:sessions_token_idx\nr:schema_migration\nr:sessions\n'
    fi
    ;;
  *pg_tables*) printf 'schema_migration\nsessions\n' ;;
  *pg_sequences*)
    if [ "$dsn" = target ] && [ "${FAKE_TARGET_SEQUENCE_DIFFERENT:-}" = yes ]; then
      printf 'sessions_id_seq:3\n'
    else
      printf 'sessions_id_seq:2\n'
    fi
    ;;
  *to_jsonb*)
    if [ "$dsn" = target ] && [ "${FAKE_TARGET_FINGERPRINT_DIFFERENT:-}" = yes ]; then
      printf '3:%032d\n' 1
    else
      printf '2:%032d\n' 1
    fi
    ;;
  *schema_migration*)
    if [ "$dsn" = target ] && [ "${FAKE_TARGET_LEDGER_DIFFERENT:-}" = yes ]; then
      printf '20210217152610:1\n'
    else
      printf '20210217152610:1\n20210813104751:1\n'
    fi
    ;;
  *) echo "unexpected query: $sql" >&2; exit 1 ;;
esac
EOF

cat >"$fake_bin/pg_dump" <<'EOF'
#!/usr/bin/env bash
schema_seen=no
for argument in "$@"; do
  case "$argument" in
    --schema=data) schema_seen=yes ;;
    --file=*) output="${argument#--file=}" ;;
  esac
done
[ "$schema_seen" = yes ]
printf 'dump' >"$output"
EOF

cat >"$fake_bin/pg_restore" <<'EOF'
#!/usr/bin/env bash
[ -f "${!#}" ]
EOF

chmod +x "$fake_bin/psql" "$fake_bin/pg_dump" "$fake_bin/pg_restore"
export PATH="$fake_bin:$PATH"
export ORY_COPY_SOURCE_DSN=source
export ORY_COPY_TARGET_DSN=target
export ORY_COPY_WRITES_PAUSED=yes

backup="$test_directory/success.dump"
"$copy_script" "$backup"
[ -s "$backup" ]
[ "$(stat -c %a "$backup")" = 600 ]

if FAKE_SAME_IDENTITY=yes "$copy_script" "$test_directory/same.dump" >"$test_directory/out" 2>&1; then
  echo "same database was accepted" >&2
  exit 1
fi
grep -q 'source and target resolve to the same database' "$test_directory/out"

if FAKE_BAD_SEARCH_PATH=yes "$copy_script" "$test_directory/search-path.dump" >"$test_directory/out" 2>&1; then
  echo "bad target search_path was accepted" >&2
  exit 1
fi
grep -q 'target role search_path does not include data' "$test_directory/out"

if FAKE_TARGET_RELATIONS=1 "$copy_script" "$test_directory/nonempty.dump" >"$test_directory/out" 2>&1; then
  echo "non-empty target was accepted" >&2
  exit 1
fi
grep -q 'target schema data is not empty' "$test_directory/out"

if FAKE_TARGET_LEDGER_DIFFERENT=yes "$copy_script" "$test_directory/ledger.dump" >"$test_directory/out" 2>&1; then
  echo "different Ory migration ledger was accepted" >&2
  exit 1
fi
grep -q 'Ory migration ledger differs after restore' "$test_directory/out"

if FAKE_TARGET_INVENTORY_DIFFERENT=yes "$copy_script" "$test_directory/inventory.dump" >"$test_directory/out" 2>&1; then
  echo "different schema inventory was accepted" >&2
  exit 1
fi
grep -q 'schema object inventory differs after restore' "$test_directory/out"

if FAKE_TARGET_SEQUENCE_DIFFERENT=yes "$copy_script" "$test_directory/sequence.dump" >"$test_directory/out" 2>&1; then
  echo "different sequence state was accepted" >&2
  exit 1
fi
grep -q 'sequence values differ after restore' "$test_directory/out"

fingerprint_backup="$test_directory/fingerprint.dump"
if FAKE_TARGET_FINGERPRINT_DIFFERENT=yes "$copy_script" "$fingerprint_backup" >"$test_directory/out" 2>&1; then
  echo "different table contents were accepted" >&2
  exit 1
fi
grep -q 'table fingerprints differ after restore' "$test_directory/out"
[ -s "$fingerprint_backup" ]

if ORY_COPY_WRITES_PAUSED=no "$copy_script" "$test_directory/unpaused.dump" >"$test_directory/out" 2>&1; then
  echo "unpaused writes were accepted" >&2
  exit 1
fi
grep -q 'pause and drain writes' "$test_directory/out"

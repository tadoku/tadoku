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
table=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --dbname=*) dsn="${1#--dbname=}" ;;
    -c) shift; sql="$1" ;;
    -v)
      shift
      case "$1" in table=*) table="${1#table=}" ;; esac
      ;;
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
  *pg_class*) echo "${FAKE_TARGET_RELATIONS:-0}" ;;
  *pg_tables*) printf 'logs\nprofiles\n' ;;
  *schema_migrations*) echo '29:false' ;;
  *to_jsonb*)
    if [ "$dsn" = target ] && [ "${FAKE_TARGET_FINGERPRINT_DIFFERENT:-}" = yes ]; then
      printf '3:%032d\n' "${#table}"
    else
      printf '2:%032d\n' "${#table}"
    fi
    ;;
  *) echo "unexpected query: $sql" >&2; exit 1 ;;
esac
EOF

cat >"$fake_bin/pg_dump" <<'EOF'
#!/usr/bin/env bash
for required_argument in --schema=data --extension=uuid-ossp; do
  found=no
  for argument in "$@"; do
    if [ "$argument" = "$required_argument" ]; then
      found=yes
      break
    fi
  done
  [ "$found" = yes ] || {
    echo "missing dump scope: $required_argument" >&2
    exit 1
  }
done
for argument in "$@"; do
  case "$argument" in --file=*) printf 'dump' >"${argument#--file=}" ;; esac
done
EOF

cat >"$fake_bin/pg_restore" <<'EOF'
#!/usr/bin/env bash
[ -f "${!#}" ]
EOF

chmod +x "$fake_bin/psql" "$fake_bin/pg_dump" "$fake_bin/pg_restore"
export PATH="$fake_bin:$PATH"
export TADOKU_COPY_SOURCE_DSN=source
export TADOKU_COPY_TARGET_DSN=target
export TADOKU_COPY_EXPECTED_VERSION=29
export TADOKU_COPY_WRITES_PAUSED=yes

backup="$test_directory/success.dump"
"$copy_script" "$backup"
[ -s "$backup" ]
[ "$(stat -c %a "$backup")" = 600 ]

if FAKE_SAME_IDENTITY=yes "$copy_script" "$test_directory/same.dump" >"$test_directory/out" 2>&1; then
  echo "same database was accepted" >&2
  exit 1
fi
grep -q 'source and target resolve to the same database' "$test_directory/out"

if FAKE_TARGET_RELATIONS=1 "$copy_script" "$test_directory/nonempty.dump" >"$test_directory/out" 2>&1; then
  echo "non-empty target was accepted" >&2
  exit 1
fi
grep -q 'target schema data is not empty' "$test_directory/out"

fingerprint_backup="$test_directory/fingerprint.dump"
if FAKE_TARGET_FINGERPRINT_DIFFERENT=yes "$copy_script" "$fingerprint_backup" >"$test_directory/out" 2>&1; then
  echo "different table contents were accepted" >&2
  exit 1
fi
grep -q 'table fingerprints differ after restore' "$test_directory/out"
[ -s "$fingerprint_backup" ]

if TADOKU_COPY_WRITES_PAUSED=no "$copy_script" "$test_directory/unpaused.dump" >"$test_directory/out" 2>&1; then
  echo "unpaused writes were accepted" >&2
  exit 1
fi
grep -q 'pause and drain writes' "$test_directory/out"

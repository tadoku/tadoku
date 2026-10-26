#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_NAME="tadoku-dev-db"
DB_NAMESPACE="${TADOKU_DEV_NAMESPACE:-tdk-dev-data}"
KUBE_CONTEXT="${TADOKU_DEV_CONTEXT:-homelab-talos-dev}"
ADMIN_EMAIL="${TADOKU_DEV_ADMIN_EMAIL:-dev@tadoku.app}"
ADMIN_PASSWORD="${TADOKU_DEV_ADMIN_PASSWORD:-tadoku}"
READER_EMAIL="${TADOKU_DEV_READER_EMAIL:-reader@tadoku.app}"
READER_PASSWORD="${TADOKU_DEV_READER_PASSWORD:-tadoku}"
ADMIN_CONTEST_ID="00000000-0000-4000-8000-000000000101"
READER_CONTEST_ID="00000000-0000-4000-8000-000000000102"
NATIVE_ADMIN_ID="00000000-0000-4000-8000-000000000001"
NATIVE_READER_ID="00000000-0000-4000-8000-000000000002"

if [ "$KUBE_CONTEXT" != "homelab-talos-dev" ]; then
  echo "refusing to seed: only the homelab-talos-dev Kubernetes context is supported" >&2
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

db_pod() {
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" get pod \
    -l "application=spilo,cluster-name=${DB_NAME},spilo-role=master" \
    -o jsonpath='{.items[0].metadata.name}'
}

wait_for_db() {
  echo "waiting for ${DB_NAME} pod..."
  local _i found=""
  for _i in $(seq 1 60); do
    if [ -n "$(kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" get pod \
      -l "application=spilo,cluster-name=${DB_NAME}" \
      -o name 2>/dev/null)" ]; then
      found=1
      break
    fi
    sleep 5
  done
  if [ -z "$found" ]; then
    echo "timed out waiting for ${DB_NAME} pod to appear in namespace ${DB_NAMESPACE}; is the dev Postgres cluster deployed?" >&2
    exit 1
  fi
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" wait \
    --for=condition=Ready \
    pod \
    -l "application=spilo,cluster-name=${DB_NAME}" \
    --timeout=300s
}

database_password() {
  local user="$1"
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" get secret \
    "${user}.${DB_NAME}.credentials.postgresql.acid.zalan.do" \
    -o jsonpath='{.data.password}' | base64 --decode
}

wait_for_relation() {
  local database="$1"
  local user="$2"
  local relation="$3"
  local pod
  local database_password_value
  pod="$(db_pod)"
  database_password_value="$(database_password "$user")"

  for _ in $(seq 1 120); do
    if kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" exec "$pod" -- env PGPASSWORD="$database_password_value" PGSSLMODE=require \
      psql -X -qAt -h "${DB_NAME}.${DB_NAMESPACE}" -U "$user" -d "$database" \
      -c "select coalesce(to_regclass('public.${relation}') is not null, false)" 2>/dev/null | grep -q '^t$'; then
      return 0
    fi
    sleep 2
  done

  echo "timed out waiting for ${database}.${relation}; migrations may not have completed" >&2
  exit 1
}

native_fixture_ids() {
  local pod
  local database_password_value
  pod="$(db_pod)"
  database_password_value="$(database_password tadoku)"
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$database_password_value" PGSSLMODE=require \
    psql -X -q -At \
      -v ON_ERROR_STOP=1 \
      -v "admin_contest_id=${ADMIN_CONTEST_ID}" \
      -v "reader_contest_id=${READER_CONTEST_ID}" \
      -v "native_admin_id=${NATIVE_ADMIN_ID}" \
      -v "native_reader_id=${NATIVE_READER_ID}" \
      -h "${DB_NAME}.${DB_NAMESPACE}" -U tadoku -d tadoku <<'SQL'
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true) as tenant \gset
select
  coalesce(
    (select owner_user_id from contests where tenant = 'tadoku/prod' and id = :'admin_contest_id'),
    :'native_admin_id'
  ) as admin_user_id,
  coalesce(
    (select owner_user_id from contests where tenant = 'tadoku/prod' and id = :'reader_contest_id'),
    :'native_reader_id'
  ) as reader_user_id
\gset
\echo :admin_user_id :reader_user_id
commit;
SQL
}

NATIVE_RUNNER=""
NATIVE_FORWARD_LOG=""
NATIVE_FORWARD_PID=""

# Stops the owned port-forward and removes the runner and its log; safe to run more than once.
cleanup_native_accounts() {
  if [ -n "$NATIVE_FORWARD_PID" ]; then
    kill "$NATIVE_FORWARD_PID" 2>/dev/null || true
    wait "$NATIVE_FORWARD_PID" 2>/dev/null || true
  fi
  rm -f -- ${NATIVE_RUNNER:+"$NATIVE_RUNNER"} ${NATIVE_FORWARD_LOG:+"$NATIVE_FORWARD_LOG"}
  NATIVE_RUNNER=""
  NATIVE_FORWARD_LOG=""
  NATIVE_FORWARD_PID=""
}

# Cleans up, then ends the seed with the signal it received.
interrupt_native_accounts() {
  cleanup_native_accounts
  trap - "$1"
  kill -"$1" "$$"
}

seed_native_accounts() {
  local admin_id="$1"
  local reader_id="$2"
  local pod
  local port=""
  local database_password_value
  trap cleanup_native_accounts EXIT
  trap 'interrupt_native_accounts HUP' HUP
  trap 'interrupt_native_accounts INT' INT
  trap 'interrupt_native_accounts TERM' TERM
  database_password_value="$(database_password tadoku)"
  NATIVE_RUNNER="$(mktemp)"
  # Build before forwarding; the database credentials reach only the built command.
  (cd "$ROOT" && env -u TADOKU_DEV_ADMIN_PASSWORD -u TADOKU_DEV_READER_PASSWORD \
    bazel run --noshow_progress --script_path="$NATIVE_RUNNER" //services/tadoku-api/cmd/tadoku-tenant)
  pod="$(db_pod)"
  NATIVE_FORWARD_LOG="$(mktemp)"

  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" port-forward --address 127.0.0.1 "pod/${pod}" :5432 \
    > "$NATIVE_FORWARD_LOG" 2>&1 &
  NATIVE_FORWARD_PID=$!
  for _ in $(seq 1 100); do
    port="$(sed -n 's/^Forwarding from 127\.0\.0\.1:\([0-9][0-9]*\) -> 5432$/\1/p' "$NATIVE_FORWARD_LOG")"
    if [ -n "$port" ] || ! kill -0 "$NATIVE_FORWARD_PID" 2>/dev/null; then
      break
    fi
    sleep 0.2
  done
  if [ -z "$port" ]; then
    echo "could not forward the development database to loopback" >&2
    exit 1
  fi

  local status=0
  TENANT_PRODUCTION=false \
    TENANT_POSTGRES_HOST=127.0.0.1 \
    TENANT_POSTGRES_PORT="$port" \
    TENANT_POSTGRES_DATABASE=tadoku \
    TENANT_POSTGRES_USER=tadoku \
    TENANT_POSTGRES_PASSWORD="$database_password_value" \
    TENANT_POSTGRES_SSLMODE=require \
    TADOKU_DEV_ADMIN_EMAIL="$ADMIN_EMAIL" \
    TADOKU_DEV_ADMIN_PASSWORD="$ADMIN_PASSWORD" \
    TADOKU_DEV_READER_EMAIL="$READER_EMAIL" \
    TADOKU_DEV_READER_PASSWORD="$READER_PASSWORD" \
    "$NATIVE_RUNNER" canonical-development-accounts --admin-id "$admin_id" --reader-id "$reader_id" || status=$?
  cleanup_native_accounts
  trap - EXIT HUP INT TERM
  if [ "$status" -ne 0 ]; then
    echo "native fixture accounts were not seeded; nothing else was changed" >&2
    exit "$status"
  fi
  echo "seeded native accounts ${admin_id} ${reader_id}"
}

seed_admin() {
  local subject_id="$1"
  local pod
  local database_password_value
  pod="$(db_pod)"
  database_password_value="$(database_password tadoku)"

  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$database_password_value" PGSSLMODE=require \
    psql -X -q \
      -v ON_ERROR_STOP=1 \
      -v tenant=tadoku/prod \
      -v "admin_user_id=${subject_id}" \
      -h "${DB_NAME}.${DB_NAMESPACE}" -U tadoku -d tadoku <<'SQL'
begin;
select set_config('tadoku.tenant', :'tenant', true) as tenant \gset
insert into permission_relationships (tenant, object_type, object_id, relation, subject_type, subject_id)
values (:'tenant', 'app', 'root', 'admins', 'user', lower(:'admin_user_id'))
on conflict do nothing;
commit;
SQL
  echo "seeded administrator ${subject_id}"
}

run_seed_sql() {
  local database="$1"
  local user="$2"
  local file="$3"
  local pod
  pod="$(db_pod)"

  echo "seeding ${database} from ${file#"$ROOT"/}"
  local database_password_value
  database_password_value="$(database_password "$user")"
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$database_password_value" PGSSLMODE=require \
    psql -X \
      -v ON_ERROR_STOP=1 \
      -v tenant=tadoku/prod \
      -v "admin_user_id=${ADMIN_USER_ID}" \
      -v "reader_user_id=${READER_USER_ID}" \
      -h "${DB_NAME}.${DB_NAMESPACE}" -U "$user" -d "$database" < "$file"
}

require_cmd kubectl
require_cmd bazel
cluster_server="$(kubectl --context "$KUBE_CONTEXT" config view --minify --output=jsonpath='{.clusters[0].cluster.server}')"
if [ "$cluster_server" != "https://omni.lab:8100" ]; then
  echo "refusing to seed: homelab-talos-dev must point to https://omni.lab:8100" >&2
  exit 1
fi
wait_for_db

wait_for_relation tadoku tadoku users
wait_for_relation tadoku tadoku accounts
wait_for_relation tadoku tadoku contests
FIXTURE_IDS="$(native_fixture_ids)"
read -r ADMIN_USER_ID READER_USER_ID <<< "$FIXTURE_IDS"
echo "seeding native accounts..."
seed_native_accounts "$ADMIN_USER_ID" "$READER_USER_ID"
export ADMIN_USER_ID READER_USER_ID
wait_for_relation tadoku tadoku permission_relationships
seed_admin "$ADMIN_USER_ID"

wait_for_relation tadoku tadoku pages

run_seed_sql tadoku tadoku "$ROOT/scripts/dev/seed/immersion.sql"
run_seed_sql tadoku tadoku "$ROOT/scripts/dev/seed/content.sql"

echo "dev seed complete"
echo "admin: ${ADMIN_EMAIL}"
echo "reader: ${READER_EMAIL}"

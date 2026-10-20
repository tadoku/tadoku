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
# The base host always serves the canonical tenant, so its anonymous session names that tenant's identity provider.
SESSION_URL="https://tadoku.dev.lab/api/internal/identity/session"
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

identity_provider() {
  local body
  if ! body="$(curl --fail --silent --show-error --max-time 15 --proto =https \
    -H 'Accept: application/json' "$SESSION_URL")"; then
    echo "refusing to seed: cannot read the canonical identity provider from ${SESSION_URL}" >&2
    return 1
  fi
  case "$body" in
    '{"account":null,"account_app":"v1"}') echo v1 ;;
    '{"account":null,"account_app":"v2"}') echo v2 ;;
    *)
      echo "refusing to seed: unexpected identity session response from ${SESSION_URL}" >&2
      return 1
      ;;
  esac
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

seed_native_accounts() {
  local admin_id="$1"
  local reader_id="$2"
  local pod
  local forward_log
  local forward_pid
  local port=""
  local runner
  local database_password_value
  database_password_value="$(database_password tadoku)"
  runner="$(mktemp)"
  # Build before forwarding; the database credentials reach only the built command.
  (cd "$ROOT" && env -u TADOKU_DEV_ADMIN_PASSWORD -u TADOKU_DEV_READER_PASSWORD \
    bazel run --noshow_progress --script_path="$runner" //services/tadoku-api/cmd/tadoku-tenant)
  pod="$(db_pod)"
  forward_log="$(mktemp)"

  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" port-forward --address 127.0.0.1 "pod/${pod}" :5432 \
    > "$forward_log" 2>&1 &
  forward_pid=$!
  for _ in $(seq 1 100); do
    port="$(sed -n 's/^Forwarding from 127\.0\.0\.1:\([0-9][0-9]*\) -> 5432$/\1/p' "$forward_log")"
    if [ -n "$port" ] || ! kill -0 "$forward_pid" 2>/dev/null; then
      break
    fi
    sleep 0.2
  done
  if [ -z "$port" ]; then
    kill "$forward_pid" 2>/dev/null || true
    rm -f "$forward_log" "$runner"
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
    "$runner" canonical-development-accounts --admin-id "$admin_id" --reader-id "$reader_id" || status=$?
  kill "$forward_pid" 2>/dev/null || true
  wait "$forward_pid" 2>/dev/null || true
  rm -f "$forward_log" "$runner"
  if [ "$status" -ne 0 ]; then
    echo "native fixture accounts were not seeded; nothing else was changed" >&2
    exit "$status"
  fi
  echo "seeded native accounts ${admin_id} ${reader_id}"
}

seed_identity() {
  local email="$1"
  local display_name="$2"
  local password="$3"
  local pod_name
  pod_name="tadoku-dev-identity-seed-${RANDOM}-${RANDOM}"

  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" run "$pod_name" \
    --quiet \
    --rm \
    -i \
    --restart=Never \
    --image=python:3.12-alpine \
    --env="SEED_EMAIL=${email}" \
    --env="SEED_DISPLAY_NAME=${display_name}" \
    --env="SEED_PASSWORD=${password}" \
    --env="KRATOS_ADMIN_URL=${TADOKU_KRATOS_ADMIN_URL:-http://kratos-admin.tdk-dev-kratos}" \
    -- python - <<'PY'
import json
import os
import sys
import time
import urllib.error
import urllib.request

kratos_admin = os.getenv("KRATOS_ADMIN_URL", "http://kratos-admin").rstrip("/")
email = os.environ["SEED_EMAIL"]
display_name = os.environ["SEED_DISPLAY_NAME"]
password = os.environ["SEED_PASSWORD"]
SEED_MARKER = "tadoku-dev-seed"

def request(method, path, payload=None):
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(f"{kratos_admin}{path}", data=data, method=method)
    req.add_header("Accept", "application/json")
    if payload is not None:
        req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req, timeout=10) as resp:
        body = resp.read().decode("utf-8")
        return json.loads(body) if body else {}

def find_identity():
    for page in range(1, 51):
        identities = request("GET", f"/admin/identities?per_page=250&page={page}")
        if not identities:
            return None
        for identity in identities:
            traits = identity.get("traits") or {}
            if traits.get("email") == email:
                return identity
    return None

deadline = time.time() + 180
last_error = None
while time.time() < deadline:
    try:
        identity = find_identity()
        if identity:
            metadata = identity.get("metadata_admin") or {}
            if metadata.get("seeded_by") != SEED_MARKER:
                print(
                    f"error: identity {email} already exists but is not owned by the dev seed "
                    f"(missing metadata_admin.seeded_by={SEED_MARKER}); refusing to touch its credentials.\n"
                    "remediation: inspect the conflicting identity; only after explicit approval, "
                    "remove that disposable identity and rerun `make dev-seed`. Do not reset shared databases "
                    "or add a seed ownership marker to an unrelated identity.",
                    file=sys.stderr,
                )
                sys.exit(2)
            print(identity["id"])
            sys.exit(0)

        identity = request("POST", "/admin/identities", {
            "schema_id": "user",
            "traits": {
                "email": email,
                "display_name": display_name,
            },
            "metadata_admin": {
                "seeded_by": SEED_MARKER,
            },
            "credentials": {
                "password": {
                    "config": {
                        "password": password,
                    },
                },
            },
        })
        print(identity["id"])
        sys.exit(0)
    except Exception as exc:
        last_error = exc
        time.sleep(2)

print(f"failed to seed kratos identity for {email}: {last_error}", file=sys.stderr)
sys.exit(1)
PY
}

refresh_identity_password() {
  local email="$1"
  local password="$2"
  local pod
  pod="$(db_pod)"

  local database_password_value
  database_password_value="$(database_password kratos)"
  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$database_password_value" PGSSLMODE=require \
    psql -X -q \
      -v ON_ERROR_STOP=1 \
      -v "seed_email=${email}" \
      -v "seed_password=${password}" \
      -h "${DB_NAME}.${DB_NAMESPACE}" -U kratos -d kratos <<'SQL'
create extension if not exists pgcrypto;
update identity_credentials ic
set config = jsonb_set(
      ic.config,
      '{hashed_password}',
      to_jsonb(crypt(:'seed_password', gen_salt('bf', 12)))
    ),
    updated_at = now()
from identities i,
     identity_credential_types ict
where ic.identity_id = i.id
  and ic.identity_credential_type_id = ict.id
  and ict.name = 'password'
  and i.traits->>'email' = :'seed_email'
  and i.metadata_admin->>'seeded_by' = 'tadoku-dev-seed'
  and not coalesce(
    ic.config->>'hashed_password' like '$2a$%'
      and crypt(:'seed_password', ic.config->>'hashed_password') = ic.config->>'hashed_password',
    false
  );
SQL
  echo "checked password for seeded identity ${email}"
}

seed_keto_admin() {
  local subject_id="$1"
  local pod_name
  pod_name="tadoku-dev-keto-seed-${RANDOM}-${RANDOM}"

  kubectl --context "$KUBE_CONTEXT" -n "$DB_NAMESPACE" run "$pod_name" \
    --quiet \
    --rm \
    -i \
    --restart=Never \
    --image=python:3.12-alpine \
    --env="SEED_SUBJECT_ID=${subject_id}" \
    --env="KETO_WRITE_URL=${TADOKU_KETO_WRITE_URL:-http://keto-write.tdk-dev-keto:4467}" \
    --env="KETO_READ_URL=${TADOKU_KETO_READ_URL:-http://keto-read.tdk-dev-keto:4466}" \
    -- python - <<'PY'
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

keto_write = os.getenv("KETO_WRITE_URL", "http://keto-write:4467").rstrip("/")
keto_read = os.getenv("KETO_READ_URL", "http://keto-read:4466").rstrip("/")
subject_id = os.environ["SEED_SUBJECT_ID"]
payload = {
    "namespace": os.getenv("KETO_NAMESPACE", "app"),
    "object": os.getenv("KETO_OBJECT", "tadoku"),
    "relation": os.getenv("KETO_RELATION", "admins"),
    "subject_id": subject_id,
}

deadline = time.time() + 180
last_error = None
while time.time() < deadline:
    try:
        with urllib.request.urlopen(
            f"{keto_read}/relation-tuples?{urllib.parse.urlencode(payload)}",
            timeout=10,
        ) as resp:
            if json.load(resp).get("relation_tuples"):
                print(f"keto admin already seeded for {subject_id}")
                sys.exit(0)
    except Exception as exc:
        last_error = exc
        time.sleep(2)
        continue
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        f"{keto_write}/admin/relation-tuples",
        data=data,
        method="PUT",
    )
    req.add_header("Accept", "application/json")
    req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            if resp.getcode() in (200, 201, 204):
                print(f"seeded keto admin for {subject_id}")
                sys.exit(0)
    except urllib.error.HTTPError as exc:
        if exc.code == 409:
            print(f"keto admin already seeded for {subject_id}")
            sys.exit(0)
        last_error = f"http {exc.code}: {exc.read().decode('utf-8', 'replace')}"
    except Exception as exc:
        last_error = exc
    time.sleep(2)

print(f"failed to seed keto admin for {subject_id}: {last_error}", file=sys.stderr)
sys.exit(1)
PY
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
      -v "seed_users=${SEED_USERS}" \
      -v "admin_user_id=${ADMIN_USER_ID}" \
      -v "reader_user_id=${READER_USER_ID}" \
      -h "${DB_NAME}.${DB_NAMESPACE}" -U "$user" -d "$database" < "$file"
}

require_cmd kubectl
require_cmd curl
cluster_server="$(kubectl --context "$KUBE_CONTEXT" config view --minify --output=jsonpath='{.clusters[0].cluster.server}')"
if [ "$cluster_server" != "https://omni.lab:8100" ]; then
  echo "refusing to seed: homelab-talos-dev must point to https://omni.lab:8100" >&2
  exit 1
fi
PROVIDER="$(identity_provider)"
echo "canonical identity provider: ${PROVIDER}"
wait_for_db

if [ "$PROVIDER" = v2 ]; then
  require_cmd bazel
  wait_for_relation tadoku tadoku users
  wait_for_relation tadoku tadoku accounts
  wait_for_relation tadoku tadoku contests
  FIXTURE_IDS="$(native_fixture_ids)"
  read -r ADMIN_USER_ID READER_USER_ID <<< "$FIXTURE_IDS"
  echo "seeding native accounts..."
  seed_native_accounts "$ADMIN_USER_ID" "$READER_USER_ID"
  SEED_USERS=false
else
  echo "seeding kratos identities..."
  ADMIN_USER_ID="$(seed_identity "$ADMIN_EMAIL" "Dev Admin" "$ADMIN_PASSWORD" | tail -n 1)"
  READER_USER_ID="$(seed_identity "$READER_EMAIL" "Dev Reader" "$READER_PASSWORD" | tail -n 1)"
  refresh_identity_password "$ADMIN_EMAIL" "$ADMIN_PASSWORD"
  refresh_identity_password "$READER_EMAIL" "$READER_PASSWORD"
  SEED_USERS=true
fi
export ADMIN_USER_ID READER_USER_ID
seed_keto_admin "$ADMIN_USER_ID"

wait_for_relation tadoku tadoku users
wait_for_relation tadoku tadoku pages
wait_for_relation tadoku tadoku profiles

run_seed_sql tadoku tadoku "$ROOT/scripts/dev/seed/immersion.sql"
run_seed_sql tadoku tadoku "$ROOT/scripts/dev/seed/profile.sql"
run_seed_sql tadoku tadoku "$ROOT/scripts/dev/seed/content.sql"

echo "dev seed complete"
echo "admin: ${ADMIN_EMAIL}"
echo "reader: ${READER_EMAIL}"

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_NAME="tadoku-dev-db"
DB_NAMESPACE="${TADOKU_DEV_NAMESPACE:-default}"
DB_PASSWORD="${TADOKU_DEV_DB_PASSWORD:-dev-foobar}"
ADMIN_EMAIL="${TADOKU_DEV_ADMIN_EMAIL:-dev@tadoku.app}"
ADMIN_PASSWORD="${TADOKU_DEV_ADMIN_PASSWORD:-tadoku}"
READER_EMAIL="${TADOKU_DEV_READER_EMAIL:-reader@tadoku.app}"
READER_PASSWORD="${TADOKU_DEV_READER_PASSWORD:-tadoku}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

db_pod() {
  kubectl -n "$DB_NAMESPACE" get pod \
    -l "application=spilo,cluster-name=${DB_NAME},spilo-role=master" \
    -o jsonpath='{.items[0].metadata.name}'
}

wait_for_db() {
  echo "waiting for ${DB_NAME} pod..."
  local _i found=""
  for _i in $(seq 1 60); do
    if [ -n "$(kubectl -n "$DB_NAMESPACE" get pod \
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
  kubectl -n "$DB_NAMESPACE" wait \
    --for=condition=Ready \
    pod \
    -l "application=spilo,cluster-name=${DB_NAME}" \
    --timeout=300s
}

psql_url() {
  local database="$1"
  local user="$2"
  printf 'postgres://%s:%s@%s.%s/%s?sslmode=require' \
    "$user" "$DB_PASSWORD" "$DB_NAME" "$DB_NAMESPACE" "$database"
}

wait_for_relation() {
  local database="$1"
  local user="$2"
  local relation="$3"
  local pod
  pod="$(db_pod)"

  for _ in $(seq 1 120); do
    if kubectl -n "$DB_NAMESPACE" exec "$pod" -- env PGPASSWORD="$DB_PASSWORD" \
      psql -X -qAt "$(psql_url "$database" "$user")" \
      -c "select coalesce(to_regclass('public.${relation}') is not null, false)" 2>/dev/null | grep -q '^t$'; then
      return 0
    fi
    sleep 2
  done

  echo "timed out waiting for ${database}.${relation}; migrations may not have completed" >&2
  exit 1
}

seed_identity() {
  local email="$1"
  local display_name="$2"
  local password="$3"
  local pod_name
  pod_name="tadoku-dev-identity-seed-${RANDOM}-${RANDOM}"

  kubectl -n "$DB_NAMESPACE" run "$pod_name" \
    --quiet \
    --rm \
    -i \
    --restart=Never \
    --image=python:3.12-alpine \
    --env="SEED_EMAIL=${email}" \
    --env="SEED_DISPLAY_NAME=${display_name}" \
    --env="SEED_PASSWORD=${password}" \
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
                print(f"identity {email} not owned by dev seed; leaving credentials untouched", file=sys.stderr)
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

  kubectl -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$DB_PASSWORD" \
    psql -X -q \
      -v ON_ERROR_STOP=1 \
      -v "seed_email=${email}" \
      -v "seed_password=${password}" \
      "$(psql_url kratos kratos)" <<'SQL'
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
  and i.metadata_admin->>'seeded_by' = 'tadoku-dev-seed';
SQL
  echo "refreshed password for seeded identity ${email} (no-op if not seed-owned)"
}

seed_keto_admin() {
  local subject_id="$1"
  local pod_name
  pod_name="tadoku-dev-keto-seed-${RANDOM}-${RANDOM}"

  kubectl -n "$DB_NAMESPACE" run "$pod_name" \
    --quiet \
    --rm \
    -i \
    --restart=Never \
    --image=python:3.12-alpine \
    --env="SEED_SUBJECT_ID=${subject_id}" \
    -- python - <<'PY'
import json
import os
import sys
import time
import urllib.error
import urllib.request

keto_write = os.getenv("KETO_WRITE_URL", "http://keto-write:4467").rstrip("/")
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
  kubectl -n "$DB_NAMESPACE" exec -i "$pod" -- env PGPASSWORD="$DB_PASSWORD" \
    psql -X \
      -v ON_ERROR_STOP=1 \
      -v "admin_user_id=${ADMIN_USER_ID}" \
      -v "reader_user_id=${READER_USER_ID}" \
      "$(psql_url "$database" "$user")" < "$file"
}

require_cmd kubectl
wait_for_db

echo "seeding kratos identities..."
ADMIN_USER_ID="$(seed_identity "$ADMIN_EMAIL" "Dev Admin" "$ADMIN_PASSWORD" | tail -n 1)"
READER_USER_ID="$(seed_identity "$READER_EMAIL" "Dev Reader" "$READER_PASSWORD" | tail -n 1)"
export ADMIN_USER_ID READER_USER_ID
refresh_identity_password "$ADMIN_EMAIL" "$ADMIN_PASSWORD"
refresh_identity_password "$READER_EMAIL" "$READER_PASSWORD"
seed_keto_admin "$ADMIN_USER_ID"

wait_for_relation immersion immersion users
wait_for_relation content content pages
wait_for_relation profile profile profiles

run_seed_sql immersion immersion "$ROOT/scripts/dev/seed/immersion.sql"
run_seed_sql profile profile "$ROOT/scripts/dev/seed/profile.sql"
run_seed_sql content content "$ROOT/scripts/dev/seed/content.sql"

echo "dev seed complete"
echo "admin: ${ADMIN_EMAIL} / ${ADMIN_PASSWORD}"
echo "reader: ${READER_EMAIL} / ${READER_PASSWORD}"

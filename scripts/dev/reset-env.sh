#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_NAME="tadoku-dev-db"
DB_NAMESPACE="${TADOKU_DEV_NAMESPACE:-default}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

delete_if_present() {
  local description="$1"
  shift
  echo "deleting ${description}..."
  "$@" --ignore-not-found=true
}

rollout_restart_if_present() {
  local namespace="$1"
  local deployment="$2"
  if kubectl -n "$namespace" get deployment "$deployment" >/dev/null 2>&1; then
    kubectl -n "$namespace" rollout restart "deployment/${deployment}"
  fi
}

rollout_wait_if_present() {
  local namespace="$1"
  local deployment="$2"
  if kubectl -n "$namespace" get deployment "$deployment" >/dev/null 2>&1; then
    kubectl -n "$namespace" rollout status "deployment/${deployment}" --timeout=300s
  fi
}

tilt_config_value() {
  local key="$1"
  local file
  for file in "$ROOT/tilt_config.json" "$ROOT/tilt_config.json.example"; do
    if [ -f "$file" ]; then
      sed -n 's/.*"'"$key"'"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$file" | head -n 1
      return 0
    fi
  done
}

wait_for_db_pod() {
  echo "waiting for ${DB_NAME} pod to be created by the operator..."
  local _i
  for _i in $(seq 1 60); do
    if [ -n "$(kubectl -n "$DB_NAMESPACE" get pod \
      -l "application=spilo,cluster-name=${DB_NAME}" \
      -o name 2>/dev/null)" ]; then
      return 0
    fi
    sleep 5
  done
  echo "timed out waiting for the operator to create the ${DB_NAME} pod" >&2
  exit 1
}

require_cmd kubectl

SHARED_CONTEXT="${TADOKU_SHARED_K8S_CONTEXT:-$(tilt_config_value shared_k8s_context)}"
SHARED_CONTEXT="${SHARED_CONTEXT:-dev-lab}"
LOCAL_CONTEXT="${TADOKU_LOCAL_K8S_CONTEXT:-$(tilt_config_value local_k8s_context)}"
LOCAL_CONTEXT="${LOCAL_CONTEXT:-orbstack}"
CURRENT_CONTEXT="$(kubectl config current-context)"

if [ "$CURRENT_CONTEXT" != "$SHARED_CONTEXT" ] && [ "$CURRENT_CONTEXT" != "$LOCAL_CONTEXT" ]; then
  echo "refusing to reset: kubectl context '${CURRENT_CONTEXT}' is not a known dev context ('${SHARED_CONTEXT}' or '${LOCAL_CONTEXT}')" >&2
  echo "switch contexts or set TADOKU_SHARED_K8S_CONTEXT / TADOKU_LOCAL_K8S_CONTEXT if your dev cluster uses a different name" >&2
  exit 1
fi

if [ "$CURRENT_CONTEXT" = "$SHARED_CONTEXT" ]; then
  if [ ! -t 0 ]; then
    echo "refusing to reset the shared dev cluster '${CURRENT_CONTEXT}' without an interactive confirmation" >&2
    exit 1
  fi
  printf "this deletes the '%s' database and its data on the shared dev cluster '%s'. type the context name to continue: " "$DB_NAME" "$CURRENT_CONTEXT"
  read -r confirmation
  if [ "$confirmation" != "$CURRENT_CONTEXT" ]; then
    echo "confirmation did not match; aborting" >&2
    exit 1
  fi
fi

echo "resetting Tadoku dev environment in context: ${CURRENT_CONTEXT}"

delete_if_present "postgres operator cluster ${DB_NAME}" \
  kubectl -n "$DB_NAMESPACE" delete "postgresql.acid.zalan.do/${DB_NAME}" --wait=true

kubectl -n "$DB_NAMESPACE" wait \
  --for=delete pod \
  -l "application=spilo,cluster-name=${DB_NAME}" \
  --timeout=180s >/dev/null 2>&1 || true

delete_if_present "postgres persistent volumes for ${DB_NAME}" \
  kubectl -n "$DB_NAMESPACE" delete pvc -l "application=spilo,cluster-name=${DB_NAME}" --wait=true

echo "recreating operator-managed postgres cluster..."
kubectl apply -f "$ROOT/k8s/dev/postgres.yaml"
wait_for_db_pod
kubectl -n "$DB_NAMESPACE" wait \
  --for=condition=Ready \
  pod \
  -l "application=spilo,cluster-name=${DB_NAME}" \
  --timeout=300s

echo "restarting services so startup migrations run against the fresh database..."
rollout_restart_if_present default kratos
rollout_restart_if_present default keto-read
rollout_restart_if_present default keto-write
rollout_restart_if_present tdk-authz-api authz-api
rollout_restart_if_present tdk-content-api content-api
rollout_restart_if_present tdk-immersion-api immersion-api
rollout_restart_if_present tdk-profile-api profile-api

rollout_wait_if_present default kratos
rollout_wait_if_present default keto-read
rollout_wait_if_present default keto-write
rollout_wait_if_present tdk-authz-api authz-api
rollout_wait_if_present tdk-content-api content-api
rollout_wait_if_present tdk-immersion-api immersion-api
rollout_wait_if_present tdk-profile-api profile-api

"$ROOT/scripts/dev/seed-db.sh"

echo "dev environment reset complete"

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

require_cmd kubectl

echo "resetting Tadoku dev environment in context: $(kubectl config current-context)"

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

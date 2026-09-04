#!/usr/bin/env bash
set -euo pipefail

DB_NAME="tadoku-dev-db"
DB_NAMESPACE="${TADOKU_DEV_NAMESPACE:-default}"
KUBECTL_CONTEXT="${1:-}"
kubectl_args=()
if [ -n "$KUBECTL_CONTEXT" ]; then
  kubectl_args+=(--context "$KUBECTL_CONTEXT")
fi

kube() {
  kubectl "${kubectl_args[@]}" "$@"
}

wait_for_secret() {
  local secret_name="$1"
  local attempt

  for attempt in $(seq 1 60); do
    if kube -n "$DB_NAMESPACE" get secret "$secret_name" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  echo "timed out waiting for operator-managed Secret ${DB_NAMESPACE}/${secret_name}" >&2
  return 1
}

sync_secret() {
  local database_user="$1"
  local target_namespace="$2"
  local secret_name="${database_user}.${DB_NAME}.credentials.postgresql.acid.zalan.do"
  local owner_uid

  wait_for_secret "$secret_name"

  owner_uid="$(kube get namespace "$target_namespace" -o jsonpath='{.metadata.uid}')"
  if [ -z "$owner_uid" ]; then
    echo "missing target namespace ${target_namespace}" >&2
    return 1
  fi

  kube -n "$DB_NAMESPACE" get secret "$secret_name" \
    -o jsonpath='{.data.username}' | base64 --decode > "$temp_dir/username"
  kube -n "$DB_NAMESPACE" get secret "$secret_name" \
    -o jsonpath='{.data.password}' | base64 --decode > "$temp_dir/password"

  kube -n "$target_namespace" create secret generic "$secret_name" \
    --from-file=username="$temp_dir/username" \
    --from-file=password="$temp_dir/password" \
    --dry-run=client \
    -o yaml | kube apply -f -

  kube -n "$target_namespace" patch secret "$secret_name" \
    --type=merge \
    --patch "{\"metadata\":{\"ownerReferences\":[{\"apiVersion\":\"v1\",\"kind\":\"Namespace\",\"name\":\"${target_namespace}\",\"uid\":\"${owner_uid}\"}]}}" \
    >/dev/null
}

if ! command -v kubectl >/dev/null 2>&1; then
  echo "missing required command: kubectl" >&2
  exit 1
fi

if ! command -v base64 >/dev/null 2>&1; then
  echo "missing required command: base64" >&2
  exit 1
fi

umask 077
temp_dir="$(mktemp -d /tmp/tadoku-db-secrets.XXXXXX)"
cleanup() {
  case "$temp_dir" in
    /tmp/tadoku-db-secrets.*) rm -rf -- "$temp_dir" ;;
    *) echo "refusing to remove unexpected temporary path: $temp_dir" >&2 ;;
  esac
}
trap cleanup EXIT

sync_secret immersion tdk-authz-api
sync_secret immersion tdk-content-api
sync_secret immersion tdk-immersion-api
sync_secret immersion tdk-profile-api

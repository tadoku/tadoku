#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
env_file="${TADOKU_DEV_KUBECONFIG_ENV:-${script_dir}/.env.local}"
if [[ -f "${env_file}" ]]; then
  env_vars=(
    TADOKU_DEV_K3S_HOST
    TADOKU_DEV_K3S_SSH_TARGET
    TADOKU_DEV_K8S_CONTEXT
    TADOKU_SHARED_K8S_CONTEXT
    TADOKU_DEV_K3S_READ_CMD
    TADOKU_DEV_K3S_TLS_SERVER_NAME
    TADOKU_DEV_KUBECONFIG
  )
  preset=()
  for name in "${env_vars[@]}"; do
    if [[ -n "${!name+x}" ]]; then
      preset+=("${name}=${!name}")
    fi
  done
  # shellcheck source=/dev/null
  source "${env_file}"
  for entry in ${preset[@]+"${preset[@]}"}; do
    declare "${entry}"
  done
fi

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    printf '%s is required; set it in the environment or %s\n' "${name}" "${env_file}" >&2
    exit 1
  fi
}

require_env TADOKU_DEV_K3S_HOST
require_env TADOKU_DEV_K3S_SSH_TARGET

context="${TADOKU_DEV_K8S_CONTEXT:-${TADOKU_SHARED_K8S_CONTEXT:-}}"
if [[ -z "${context}" ]]; then
  printf 'TADOKU_DEV_K8S_CONTEXT is required; set it in the environment or %s\n' "${env_file}" >&2
  exit 1
fi

host="${TADOKU_DEV_K3S_HOST}"
server="https://${host}:6443"
ssh_target="${TADOKU_DEV_K3S_SSH_TARGET}"
read_cmd="${TADOKU_DEV_K3S_READ_CMD:-sudo cat /etc/rancher/k3s/k3s.yaml}"
tls_server_name="${TADOKU_DEV_K3S_TLS_SERVER_NAME:-${host%%.*}}"
out="${1:-${TADOKU_DEV_KUBECONFIG:-${HOME}/.kube/${context}.yaml}}"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

if ! command -v kubectl >/dev/null 2>&1; then
  printf 'kubectl is required to prepare the shared-cluster kubeconfig\n' >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
# shellcheck disable=SC2029 # read_cmd is a full remote command; client-side expansion is intended
ssh "${ssh_target}" "${read_cmd}" > "${tmp}"

current_context="$(KUBECONFIG="${tmp}" kubectl config current-context)"
if [[ "${current_context}" != "${context}" ]]; then
  KUBECONFIG="${tmp}" kubectl config rename-context "${current_context}" "${context}" >/dev/null
fi

cluster="$(KUBECONFIG="${tmp}" kubectl config view -o "jsonpath={.contexts[?(@.name==\"${context}\")].context.cluster}")"
KUBECONFIG="${tmp}" kubectl config set-cluster "${cluster}" --server="${server}" --tls-server-name="${tls_server_name}" >/dev/null
KUBECONFIG="${tmp}" kubectl config use-context "${context}" >/dev/null

install -m 600 "${tmp}" "${out}"

printf 'wrote %s\n' "${out}"
printf 'use: export KUBECONFIG=%s\n' "${out}"

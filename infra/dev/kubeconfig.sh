#!/usr/bin/env bash
set -euo pipefail

if [[ -f .env.local ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env.local
  set +a
fi

host="${TADOKU_DEV_K3S_HOST:-cluster-node.example.invalid}"
server="https://${host}:6443"
ssh_target="${TADOKU_DEV_K3S_SSH_TARGET:-user@${host}}"
read_cmd="${TADOKU_DEV_K3S_READ_CMD:-sudo cat /etc/rancher/k3s/k3s.yaml}"
tls_server_name="${TADOKU_DEV_K3S_TLS_SERVER_NAME:-${host%%.*}}"
out="${1:-${TADOKU_DEV_KUBECONFIG:-${HOME}/.kube/dev-lab.yaml}}"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

if ! command -v kubectl >/dev/null 2>&1; then
  printf 'kubectl is required to prepare the dev-lab kubeconfig\n' >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
# shellcheck disable=SC2029 # read_cmd is a full remote command; client-side expansion is intended
ssh "${ssh_target}" "${read_cmd}" > "${tmp}"

current_context="$(KUBECONFIG="${tmp}" kubectl config current-context)"
if [[ "${current_context}" != "dev-lab" ]]; then
  KUBECONFIG="${tmp}" kubectl config rename-context "${current_context}" dev-lab >/dev/null
fi

cluster="$(KUBECONFIG="${tmp}" kubectl config view -o jsonpath='{.contexts[?(@.name=="dev-lab")].context.cluster}')"
KUBECONFIG="${tmp}" kubectl config set-cluster "${cluster}" --server="${server}" --tls-server-name="${tls_server_name}" >/dev/null
KUBECONFIG="${tmp}" kubectl config use-context dev-lab >/dev/null

install -m 600 "${tmp}" "${out}"

printf 'wrote %s\n' "${out}"
printf 'use: export KUBECONFIG=%s\n' "${out}"

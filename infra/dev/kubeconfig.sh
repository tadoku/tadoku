#!/usr/bin/env bash
set -euo pipefail

host="${TADOKU_DEV_K3S_HOST:-ct200.lab}"
server="https://${host}:6443"
ssh_target="${TADOKU_DEV_K3S_SSH_TARGET:-io@${host}}"
read_cmd="${TADOKU_DEV_K3S_READ_CMD:-sudo cat /etc/rancher/k3s/k3s.yaml}"
tls_server_name="${TADOKU_DEV_K3S_TLS_SERVER_NAME:-ct200}"
out="${1:-${TADOKU_DEV_KUBECONFIG:-${HOME}/.kube/tadoku-dev.yaml}}"

tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT

if ! command -v kubectl >/dev/null 2>&1; then
  printf 'kubectl is required to prepare the tadoku-dev kubeconfig\n' >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
ssh "${ssh_target}" "${read_cmd}" > "${tmp}"

perl -0pi -e "s#server:\\s*https://[^\\s]+#server: ${server}#" "${tmp}"
install -m 600 "${tmp}" "${out}"

current_context="$(KUBECONFIG="${out}" kubectl config current-context)"
if [[ "${current_context}" != "tadoku-dev" ]]; then
  KUBECONFIG="${out}" kubectl config rename-context "${current_context}" tadoku-dev >/dev/null
fi

cluster="$(KUBECONFIG="${out}" kubectl config view -o jsonpath='{.contexts[?(@.name=="tadoku-dev")].context.cluster}')"
KUBECONFIG="${out}" kubectl config set-cluster "${cluster}" --server="${server}" --tls-server-name="${tls_server_name}" >/dev/null
KUBECONFIG="${out}" kubectl config use-context tadoku-dev >/dev/null

chmod 600 "${out}"
printf 'wrote %s\n' "${out}"
printf 'use: export KUBECONFIG=%s\n' "${out}"

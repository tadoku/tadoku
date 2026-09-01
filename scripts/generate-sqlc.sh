#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQLC_PACKAGES=(
  "services/immersion-api/storage/postgres"
  "services/content-api/storage/postgres"
  "services/profile-api/storage/postgres"
)

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd tar

case "$(uname -s)" in
  Linux) SQLC_OS="linux" ;;
  Darwin) SQLC_OS="darwin" ;;
  *)
    echo "unsupported operating system for sqlc code generation: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) SQLC_ARCH="amd64" ;;
  arm64 | aarch64) SQLC_ARCH="arm64" ;;
  *)
    echo "unsupported architecture for sqlc code generation: $(uname -m)" >&2
    exit 1
    ;;
esac

SQLC_CODEGEN_TMP="$(mktemp -d)"
trap 'rm -rf -- "$SQLC_CODEGEN_TMP"' EXIT

sqlc_binary() {
  local version="$1"
  local binary="$SQLC_CODEGEN_TMP/bin/sqlc-${version}"
  if [ ! -x "$binary" ]; then
    local archive="$SQLC_CODEGEN_TMP/sqlc-${version}.tar.gz"
    local extract_dir="$SQLC_CODEGEN_TMP/extract-${version}"
    mkdir -p "$(dirname "$binary")" "$extract_dir"
    curl --fail --silent --show-error --location \
      "https://github.com/sqlc-dev/sqlc/releases/download/v${version}/sqlc_${version}_${SQLC_OS}_${SQLC_ARCH}.tar.gz" \
      --output "$archive"
    tar -xzf "$archive" -C "$extract_dir"
    mv "$extract_dir/sqlc" "$binary"
  fi
  printf '%s\n' "$binary"
}

for package_dir in "${SQLC_PACKAGES[@]}"; do
  sqlc_version="$(sed -nE 's|^//go:generate go install github.com/kyleconroy/sqlc/cmd/sqlc@v([^[:space:]]+)$|\1|p' "$ROOT/${package_dir}/generate.go")"
  if [ -z "$sqlc_version" ]; then
    echo "could not determine the pinned sqlc version for ${package_dir}" >&2
    exit 1
  fi

  sqlc_bin="$(sqlc_binary "$sqlc_version")"
  echo "generating ${package_dir} with sqlc v${sqlc_version}"
  (cd "$ROOT/${package_dir}" && "$sqlc_bin" generate)
done

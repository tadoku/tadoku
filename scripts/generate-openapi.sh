#!/usr/bin/env bash
set -euo pipefail

API_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$API_ROOT"
# The existing generator version is pinned in go.mod. Bazel owns its toolchain.
bazel run @com_github_deepmap_oapi_codegen//cmd/oapi-codegen -- \
  -generate types,skip-prune -package openapi \
  -o "$API_ROOT/services/tadoku-api/generated/openapi/api.gen.go" \
  "$API_ROOT/services/tadoku-api/spec/openapi.yaml"

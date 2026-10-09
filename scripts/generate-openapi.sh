#!/usr/bin/env bash
set -euo pipefail

API_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_BAZEL_VERSION="$(tr -d '[:space:]' <"$API_ROOT/.bazelversion")"
OAPI_BAZEL_VERSION="$(tr -d '[:space:]' <"$API_ROOT/tools/oapi-codegen/.bazelversion")"
if [[ "$ROOT_BAZEL_VERSION" != "$OAPI_BAZEL_VERSION" ]]; then
  echo "tools/oapi-codegen/.bazelversion ($OAPI_BAZEL_VERSION) must match root .bazelversion ($ROOT_BAZEL_VERSION)" >&2
  exit 1
fi
(
  cd "$API_ROOT/tools/oapi-codegen"
  bazel run --lockfile_mode=error @com_github_oapi_codegen_oapi_codegen_v2//cmd/oapi-codegen -- \
    -config "$API_ROOT/services/tadoku-api/spec/server-codegen.yaml" \
    -o "$API_ROOT/services/tadoku-api/generated/openapi/api.gen.go" \
    "$API_ROOT/services/tadoku-api/spec/openapi.yaml"
  bazel run --lockfile_mode=error @com_github_oapi_codegen_oapi_codegen_v2//cmd/oapi-codegen -- \
    -config "$API_ROOT/services/tadoku-api/spec/callback-server-codegen.yaml" \
    -o "$API_ROOT/services/tadoku-api/generated/openapi/callback/api.gen.go" \
    "$API_ROOT/services/tadoku-api/spec/openapi.yaml"
)

#!/usr/bin/env bash
set -euo pipefail

API_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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

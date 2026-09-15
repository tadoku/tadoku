#!/usr/bin/env bash
set -euo pipefail

API_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$API_ROOT"
# The legacy DTO generator version is pinned in the root go.mod. Bazel owns its
# toolchain, and its output remains available to every retained operation.
bazel run @com_github_deepmap_oapi_codegen//cmd/oapi-codegen -- \
  -generate types,skip-prune -package openapi \
  -o "$API_ROOT/services/tadoku-api/generated/openapi/api.gen.go" \
  "$API_ROOT/services/tadoku-api/spec/openapi.yaml"

# Server generation uses an isolated v2 tool module so its parser dependencies
# cannot upgrade the legacy generator graph.
(
  cd "$API_ROOT/tools/oapi-codegen-v2"
  bazel run --lockfile_mode=error @com_github_oapi_codegen_oapi_codegen_v2//cmd/oapi-codegen -- \
    -config "$API_ROOT/services/tadoku-api/spec/server-codegen.yaml" \
    -o "$API_ROOT/services/tadoku-api/generated/openapi/server.gen.go" \
    "$API_ROOT/services/tadoku-api/spec/openapi.yaml"
)

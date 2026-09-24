#!/usr/bin/env bash
set -euo pipefail

violations=$(bazel query --lockfile_mode=error \
  'attr("visibility", "//visibility:public", kind("go_library", //services/tadoku-api/...)) union (kind("go_library", //services/tadoku-api/...) except attr("visibility", "//services/tadoku-api:|//services/tadoku-api/|//visibility:private", kind("go_library", //services/tadoku-api/...)))')

if [[ -n "$violations" ]]; then
  echo "Tadoku API Go libraries must be private or scoped to Tadoku API (Gazelle generates public libraries):" >&2
  echo "$violations" >&2
  exit 1
fi

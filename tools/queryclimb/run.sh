#!/usr/bin/env bash
set -euo pipefail
QUERY_CLIMB_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$QUERY_CLIMB_ROOT"
exec 9>/tmp/tadoku-query-climb.lock
flock 9
bazel build --repo_contents_cache=/cache/.scratch/query-climb/bazel-repos //tools/queryclimb
QUERY_CLIMB_BIN="$(bazel info --repo_contents_cache=/cache/.scratch/query-climb/bazel-repos bazel-bin 2>/dev/null)"
bazel shutdown
flock -u 9
exec 9>&-
exec "$QUERY_CLIMB_BIN"/tools/queryclimb/queryclimb_/queryclimb --root "$QUERY_CLIMB_ROOT" "$@"

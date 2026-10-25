#!/usr/bin/env bash
set -euo pipefail

valkey_violations=$(bazel query --lockfile_mode=error \
  'attr("deps", "com_github_valkey_io_valkey_go//:valkey-go", //services/tadoku-api/...) except (//services/tadoku-api/features/leaderboard:* union //services/tadoku-api/features/identity:* union //services/tadoku-api/infra/valkey:* union //services/tadoku-api/cmd/tadoku-api:* union //services/tadoku-api/cmd/tadoku-tenant:* union //services/tadoku-api/e2e:* union //services/tadoku-api/app/worker:worker_test union //services/tadoku-api/app/tenantlifecycle:tenantlifecycle_test)')
if [[ -n "$valkey_violations" ]]; then
  echo "Direct valkey-go dependencies belong only in leaderboard, identity throttling, infra/valkey, startup or E2E tests:" >&2
  echo "$valkey_violations" >&2
  exit 1
fi

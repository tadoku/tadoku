#!/usr/bin/env bash
set -euo pipefail

valkey_violations=$(bazel query --lockfile_mode=error \
  'attr("deps", "com_github_valkey_io_valkey_go//:valkey-go", //services/tadoku-api/...) except (//services/tadoku-api/features/leaderboard:* union //services/tadoku-api/infra/valkey:* union //services/tadoku-api/cmd/tadoku-api:* union //services/tadoku-api/e2e:* union //services/tadoku-api/app/worker:worker_test)')
if [[ -n "$valkey_violations" ]]; then
  echo "Direct valkey-go dependencies belong only in leaderboard, infra/valkey, startup or E2E tests:" >&2
  echo "$valkey_violations" >&2
  exit 1
fi

keto_violations=$(bazel query --lockfile_mode=error \
  'attr("deps", "//services/common/client/keto", //services/tadoku-api/...) except (//services/tadoku-api/internal/permissions:* union //services/tadoku-api/cmd/tadoku-api:* union //services/tadoku-api/e2e:*)')
if [[ -n "$keto_violations" ]]; then
  echo "Direct raw Keto client dependencies belong only in internal/permissions, startup or E2E:" >&2
  echo "$keto_violations" >&2
  exit 1
fi

#!/usr/bin/env bash
# The Webv3 release gate, run by build-frontend-webv3.yaml and by hand; see
# docs/docs/develop/ci-workflows.md#webv3-release-gate.
#
#   webv3-release-gate.sh deploy <sha256:digest> <evidence-dir>
#   webv3-release-gate.sh test <sha256:digest> <evidence-dir>
#   webv3-release-gate.sh down <evidence-dir>
#
# Run from the repository root, on a local branch (not a detached HEAD) at the
# candidate's commit, with dev-cli installed, the homelab-talos-dev kubectl
# context, the frontend dependencies installed and paper-ui built. GATE_OWNER
# selects the dev-cli owner. test needs E2E_READER_EMAIL, E2E_READER_PASSWORD,
# E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD; GATE_FORCE_FAILURE=true adds a
# failing spec for a dry run.
set -euo pipefail

owner=${GATE_OWNER:-gha}
context=homelab-talos-dev
namespace=tdk-dev-frontend-webv3
base_backend=frontend-webv3.$namespace.svc.cluster.local

fail() {
  echo "::error::$*" >&2
  exit 1
}

require_digest() {
  [[ $1 =~ ^sha256:[0-9a-f]{64}$ ]] || fail "'$1' is not an image digest."
}

# Prints the digest the development base's Webv3 runs once its rollout has
# finished and every pod runs it, or nothing.
base_digest() {
  local deployment pods spec running
  deployment="$(kubectl --context "$context" -n "$namespace" get deployment frontend-webv3 -o json)"
  pods="$(kubectl --context "$context" -n "$namespace" get pods -l app=frontend-webv3 -o json)"
  spec="$(jq -r '.spec.replicas as $replicas
    | select(.metadata.generation == .status.observedGeneration and $replicas > 0
      and ([.status.updatedReplicas, .status.readyReplicas, .status.availableReplicas, .status.replicas]
        | all(. == $replicas)))
    | .spec.template.spec.containers[] | select(.name == "frontend-webv3") | .image
    | capture("@(?<digest>sha256:[0-9a-f]{64})$").digest' <<<"$deployment")"
  running="$(jq -r '[.items[].status.containerStatuses[]? | select(.name == "frontend-webv3") | .imageID
    | capture("@(?<digest>sha256:[0-9a-f]{64})$").digest] | unique | if length == 1 then .[0] else empty end' <<<"$pods")"
  if [[ -n $spec && $spec == "$running" ]]; then
    echo "$spec"
  fi
}

# The gate needs a deployable to create the route and its tenant. Comparing
# with HEAD selects none from the commit's own changes, and admin is one the
# suite never opens, so every Webv3 request reaches the base.
deploy() {
  local candidate=$1 evidence=$2 deadline webv3 app headers route
  require_digest "$candidate"
  deadline=$((SECONDS + ${GATE_BASE_WAIT_SECONDS:-1200}))
  until [[ $(base_digest) == "$candidate" ]]; do
    if ((SECONDS >= deadline)); then
      fail "The development base never ran $candidate; it runs '$(base_digest)'."
    fi
    echo "Waiting for the development base to run $candidate; it runs '$(base_digest)'."
    sleep 20
  done
  echo "The development base runs $candidate."

  dev up --owner "$owner" --base HEAD --no-watch --service admin 2>&1 | tee "$evidence/dev-up.log"
  webv3="$(dev url --owner "$owner" --host next.tadoku.dev.lab / | tail -n 1)"
  app="$(dev url --owner "$owner" / | tail -n 1)"
  [[ $webv3 =~ ^https://([a-z0-9-]+)\.next\.tadoku\.dev\.lab/$ ]] || fail "dev url printed '$webv3'."
  route=${BASH_REMATCH[1]}
  [[ $app == "https://$route.tadoku.dev.lab/" ]] || fail "dev url printed '$app' for route $route."

  deadline=$((SECONDS + 120))
  until headers="$(curl --silent --show-error --max-time 15 --output /dev/null --dump-header - "$webv3")" &&
    grep -qixF "x-dev-backend: $base_backend"$'\r' <<<"$headers" &&
    grep -qixF "x-dev-selected: $route"$'\r' <<<"$headers"; do
    if ((SECONDS >= deadline)); then
      fail "$webv3 is not served by the development base for route $route."
    fi
    sleep 5
  done
  printf 'ROUTE=%s\nWEBV3_BASE_URL=%s\nTADOKU_API_URL=%sapi/internal\n' "$route" "${webv3%/}" "$app" >"$evidence/route.env"
  cat "$evidence/route.env"
}

test_suite() {
  local candidate=$1 evidence=$2 forced=frontend/apps/webv3/e2e/release-gate-forced-failure.spec.ts code=0 tested
  require_digest "$candidate"
  for name in E2E_READER_EMAIL E2E_READER_PASSWORD E2E_SECOND_EMAIL E2E_SECOND_PASSWORD; do
    [[ -n ${!name:-} ]] || fail "Set $name; without it the credentialed specs skip and the suite passes."
  done
  set -a
  # shellcheck source=/dev/null
  . "$evidence/route.env"
  set +a

  if [[ ${GATE_FORCE_FAILURE:-false} == true ]]; then
    printf '%s\n' "import { expect, test } from '@playwright/test'" "" \
      "test('the release gate dry run fails on purpose', () => {" \
      "  expect('a forced failure').toBe('a passing suite')" "})" >"$forced"
    trap 'rm -f "$forced"' EXIT
  fi
  (cd frontend && PLAYWRIGHT_OUTPUT_DIR="$evidence/playwright" timeout --kill-after=60 "${GATE_SUITE_TIMEOUT:-1800}" \
    pnpm --filter webv3 test:e2e) || code=$?

  tested="$(base_digest || true)"
  jq --null-input \
    --arg candidate "$candidate" --arg tested "$tested" --arg route "${ROUTE:?route.env names no route}" --arg revision "$(git rev-parse HEAD)" \
    --arg forced "${GATE_FORCE_FAILURE:-false}" --argjson code "$code" \
    '{candidate: $candidate, tested: $tested, route: $route, revision: $revision, forcedFailure: ($forced == "true"),
      suiteExitCode: $code, passed: ($code == 0 and $tested == $candidate)}' | tee "$evidence/result.json"
  if [[ -n ${GITHUB_OUTPUT:-} ]]; then
    echo "tested_digest=$tested" >>"$GITHUB_OUTPUT"
  fi
  ((code == 0)) || fail "The suite exited $code."
  [[ $tested == "$candidate" ]] || fail "The development base moved from $candidate to '$tested' during the suite."
}

case ${1:-} in
  deploy) deploy "$2" "$3" ;;
  test) test_suite "$2" "$3" ;;
  down) dev down --owner "$owner" 2>&1 | tee "$2/dev-down.log" ;;
  *) fail "Usage: webv3-release-gate.sh deploy|test <sha256:digest> <evidence-dir>, or down <evidence-dir>" ;;
esac

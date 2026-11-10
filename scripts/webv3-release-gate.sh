#!/usr/bin/env bash
# The Webv3 release gate, run by build-frontend-webv3.yaml and by hand; see
# docs/docs/develop/webv3-release-gate.md.
#
#   webv3-release-gate.sh deploy <sha256:digest> <evidence-dir>
#   webv3-release-gate.sh test <evidence-dir>
#   webv3-release-gate.sh verify <sha256:digest> <evidence-dir>
#   webv3-release-gate.sh down <evidence-dir>
#
# Run from the repository root, on a local branch (not a detached HEAD) at the
# candidate's commit, with dev-cli installed, the frontend dependencies
# installed and paper-ui built. deploy, verify and down use the
# homelab-talos-dev context of KUBECONFIG or your default kubeconfig. test
# never uses the cluster and needs E2E_READER_EMAIL, E2E_READER_PASSWORD,
# E2E_SECOND_EMAIL and E2E_SECOND_PASSWORD. GATE_OWNER selects the dev-cli
# owner, and GATE_FORCE_FAILURE=true adds a failing spec for a dry run.
set -euo pipefail

owner=${GATE_OWNER:-gha}
context=homelab-talos-dev
namespace=tdk-dev-frontend-webv3
base_backend=frontend-webv3.$namespace.svc.cluster.local

# Skips the suite may report: they need an environment no branch tenant has. Any
# other skip, such as one for missing credentials or data, fails the gate.
allowed_skips='[
  "indexing.spec.ts › the web host is indexable and lets crawlers in",
  "webv2-release-redirect.spec.ts › webv2 sends an account with release-webv3 to the same page on Webv3 and nothing else"
]'

fail() {
  echo "::error::$*" >&2
  exit 1
}

require_digest() {
  [[ $1 =~ ^sha256:[0-9a-f]{64}$ ]] || fail "'$1' is not an image digest."
}

use_cluster() {
  if [[ -n ${KUBECONFIG:-} && ! -r $KUBECONFIG ]]; then
    fail "KUBECONFIG names $KUBECONFIG, which is missing or unreadable."
  fi
}

# Prints the development base's Webv3 as JSON: the Deployment's generation and
# image digest, each live pod's UID and image digest, and whether the rollout
# has finished with every pod on the Deployment's digest.
base_state() {
  local deployment pods
  deployment="$(kubectl --context "$context" -n "$namespace" get deployment frontend-webv3 -o json)"
  pods="$(kubectl --context "$context" -n "$namespace" get pods -l app=frontend-webv3 -o json)"
  jq -n --argjson deployment "$deployment" --argjson pods "$pods" '
    def digest: (capture("@(?<d>sha256:[0-9a-f]{64})$").d) // null;
    $deployment as $d | $d.spec.replicas as $replicas
    | ($d.spec.template.spec.containers[] | select(.name == "frontend-webv3") | .image | digest) as $digest
    | [$pods.items[] | select(.status.phase != "Succeeded" and .status.phase != "Failed")
      | {uid: .metadata.uid, terminating: (.metadata.deletionTimestamp != null),
        digest: ([.status.containerStatuses[]? | select(.name == "frontend-webv3") | .imageID | digest][0])}]
      | sort_by(.uid) as $live
    | {generation: $d.metadata.generation, digest: $digest, pods: $live,
      settled: ($d.metadata.generation == $d.status.observedGeneration and $replicas > 0
        and ([$d.status.updatedReplicas, $d.status.readyReplicas, $d.status.availableReplicas, $d.status.replicas]
          | all(. == $replicas))
        and ($live | length) == $replicas and ($live | all(.terminating | not))
        and $digest != null and ($live | all(.digest == $digest)))}'
}

# The gate needs a deployable to create the route and its tenant. Comparing
# with HEAD selects none from the commit's own changes. tadoku-api, which also
# starts the worker, is the only deployable that is not a frontend: dev-cli
# copies a frontend overlay's sources in with kubectl exec, which the gate's
# ServiceAccount may not do. Webv3 then comes from the base, and the branch
# tenant's API and worker run the candidate commit's code.
deploy() {
  local candidate=$1 evidence=$2 deadline webv3 app headers route state
  require_digest "$candidate"
  use_cluster
  deadline=$((SECONDS + ${GATE_BASE_WAIT_SECONDS:-1200}))
  until state="$(base_state)" && jq -e --arg c "$candidate" '.settled and .digest == $c' <<<"$state" >/dev/null; do
    if ((SECONDS >= deadline)); then
      fail "The development base never settled on $candidate: ${state:-unreadable}"
    fi
    echo "Waiting for the development base to run $candidate: $(jq -c '{digest, settled}' <<<"${state:-null}")"
    sleep 20
  done
  echo "The development base runs $candidate."

  dev up --owner "$owner" --base HEAD --no-watch --service tadoku-api 2>&1 | tee "$evidence/dev-up.log"
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

  # verify requires these exact pods after the suite, so no rollout can come and go unseen.
  state="$(base_state)"
  jq -e --arg c "$candidate" '.settled and .digest == $c' <<<"$state" >/dev/null ||
    fail "The development base moved off $candidate during dev up: $state"
  jq . <<<"$state" | tee "$evidence/base-before.json"
}

test_suite() {
  local evidence=$1 forced=frontend/apps/webv3/e2e/release-gate-forced-failure.spec.ts code=0 summary
  for name in E2E_READER_EMAIL E2E_READER_PASSWORD E2E_SECOND_EMAIL E2E_SECOND_PASSWORD; do
    [[ -n ${!name:-} ]] || fail "Set $name; without it the credentialed specs skip."
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

  [[ -s $evidence/playwright/results.json ]] || fail "The suite exited $code and wrote no results."
  summary="$(jq --argjson code "$code" --argjson allowed "$allowed_skips" \
    --arg route "${ROUTE:?route.env names no route}" --arg revision "$(git rev-parse HEAD)" '
    [.. | objects | select(has("specs")) | .specs[]
      | {name: "\(.file) › \(.title)", status: [.tests[].status]}] as $tests
    | [$tests[] | select(.status | index("skipped")) | .name] as $skipped
    | {route: $route, revision: $revision, exitCode: $code, stats: .stats, skipped: $skipped,
      disallowedSkips: ($skipped - $allowed),
      passed: ($code == 0 and ($tests | length) > 0 and ($skipped - $allowed | length) == 0)}' \
    "$evidence/playwright/results.json")"
  tee "$evidence/suite.json" <<<"$summary"
  ((code == 0)) || fail "The suite exited $code."
  jq -e '.disallowedSkips == []' <<<"$summary" >/dev/null || fail "The suite skipped tests the gate requires."
  jq -e '.passed' <<<"$summary" >/dev/null || fail "The suite ran no tests."
}

verify() {
  local candidate=$1 evidence=$2 after
  require_digest "$candidate"
  use_cluster
  after="$(base_state)"
  jq . <<<"$after" >"$evidence/base-after.json"
  jq -n --arg candidate "$candidate" --slurpfile before "$evidence/base-before.json" --argjson after "$after" '
    {candidate: $candidate, before: $before[0], after: $after,
      unchanged: ($before[0].pods == $after.pods and $before[0].generation == $after.generation
        and $after.settled and $after.digest == $candidate)}' | tee "$evidence/base.json"
  jq -e '.unchanged' "$evidence/base.json" >/dev/null ||
    fail "The development base changed during the suite; see base.json."
  if [[ -n ${GITHUB_OUTPUT:-} ]]; then
    echo "tested_digest=$candidate" >>"$GITHUB_OUTPUT"
  fi
}

case ${1:-} in
  deploy) deploy "$2" "$3" ;;
  test) test_suite "$2" ;;
  verify) verify "$2" "$3" ;;
  down)
    use_cluster
    dev down --owner "$owner" 2>&1 | tee "$2/dev-down.log"
    ;;
  *) fail "Usage: webv3-release-gate.sh deploy|verify <sha256:digest> <evidence-dir>, or test|down <evidence-dir>" ;;
esac

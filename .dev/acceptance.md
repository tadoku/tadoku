# Supported development acceptance (development cluster only)

Scope: webv2, auth, admin, tadoku-api and tadoku-worker. Token-reflector is base-only.
Paper styleguide is deferred; its local pnpm workflow is independent.

Record the source/CLI revision, exact command, route, Pod UID/image, timing,
browser screenshot and result outside Git. Do not include credentials or cookies.

## Failure modes and gates

1. Bazel discovery must not select unrelated apps: query reverse dependencies
   for an auth-only file, an admin-only file and shared `packages/ui` code.
   Expect auth only, admin only, and all three frontends respectively.
2. Render `kubectl --context homelab-dev kustomize k8s/dev/base`. Validate with
   kubeconform. Auth/admin ingress must use Envoy, retain TLS and hostnames,
   and preserve the more-specific `/kratos` route to shared Kratos.
3. Base: use a fresh browser profile to log in through account.tadoku.dev.lab
   with a synthetic development account, follow its return URL, and load admin.
   Reject attempted requests to production Tadoku hosts.
4. From a fresh branch with a source-only change, run `dev doctor`, then
   `dev up --owner <unique-owner>`. Confirm only the affected app starts.
   Open `dev url --host account.tadoku.dev.lab /login` or the admin equivalent.
   Verify selected backend headers, host-only cookie and an unselected context
   still seeing base. Keep real authentication: no mocked auth or bypass.
5. Change a visible application string again while the same loop runs. Verify
   browser HMR without navigation/state loss and unchanged Pod UID/image.
   Check source deletion/dependency sync where applicable; surface sync errors.
6. Repeat for webv2 and Tadoku API with `--task migrate --task seed`. Use current
   `tadoku-<route>` databases. Verify a real branch write updates its leaderboard,
   while base and another branch remain unchanged (separate DB/cache prefixes).
7. Verify failed Go compilation preserves the old process and corrected source
   recovers. Record five warm samples including all outliers: frontend p95 ≤2s,
   backend edit-to-ready/response p95 ≤10s, alongside local/cluster pressure.
8. `dev status`, service-filtered `dev logs`, then `dev down` must stop the loop
   and remove only that owner/branch in ≤30s. Branch DB retention is intentional.
   Deleted selections fall back to base and base remains available.
9. Selecting either Tadoku API or tadoku-worker must start both under one owner,
   with the same branch database and Valkey prefix. A worker-only source edit
   must refresh only its process; an API-only edit must leave the worker Pod UID
   and process unchanged. `dev status` and `dev logs tadoku-worker` must work,
   but no worker route or URL may be created. Teardown removes both owned
   workloads and keeps the branch database and other owners untouched.

Branch selection is per hostname, not an authentication cookie. For an admin
overlay calling a Tadoku API overlay on tadoku.dev.lab, visit the CLI links for
both hosts in the same browser profile. Clearing one host does not clear others.
Run browsers with Lab CA trust; note any TLS verification bypass separately.

These gates cover the supported dev-cli workflow. Repeat
them for changes to routing or synchronization; publish CLI releases only after
their live gates pass. Live teardown always requires explicit authorization.

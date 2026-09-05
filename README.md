# Tadoku Monorepo

[![Documentation](https://img.shields.io/badge/docs-online-6969FF.svg)](https://tadoku.github.io/tadoku/)
![Build Bazel](https://github.com/tadoku/tadoku/actions/workflows/build-bazel.yaml/badge.svg)
![Build Frontend Web](https://github.com/tadoku/tadoku/actions/workflows/build-frontend-web.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)

Tadoku had a significant rewrite and the documentation hasn't been updated yet to reflect these changes.
The documentation for this repository can be found at https://tadoku.github.io/tadoku/.

## Dev Environment

Use `k8s/dev/` through the root `Tiltfile` for both shared and local clusters.
Set `TADOKU_TILT_CONFIG` to a machine-level config path so the same config works
across Git worktrees, or copy `tilt_config.json.example` to the gitignored
`tilt_config.json` for a checkout-local override. Tilt fails closed when neither
is configured.

Common commands:

```sh
make dev-up      # start Tilt
make dev-down    # stop Tilt-managed resources
make dev-reset   # delete/recreate the operator-managed dev DB, rerun migrations, and seed data
make dev-seed    # rerun idempotent seed data only
make dev-logs    # stream Tilt logs
```

The dev Postgres cluster is a Zalando `postgresql` custom resource with persistent volumes, so ordinary `tilt down`/`tilt up` keeps data.
`make dev-reset` is the destructive reset path: it deletes the operator CR and its PVCs, reapplies the CR, restarts services so init-container migrations run, then seeds deterministic dev users/content/activity data.
Seed users are `dev@tadoku.app` and `reader@tadoku.app`, both with password `tadoku`.

### Tadoku API facade rehearsal

Tilt builds and starts `tadoku-api` alongside the four legacy APIs. Both local
and shared development route existing browser API URLs and the internal
feature-flag request through it. Oathkeeper authenticates requests and issues
the same tokens, then strips `/api/internal`. Tadoku API receives domain paths
such as `/content/pages/blog`, chooses the legacy service, and forwards
`/pages/blog`. Native endpoints can replace those proxy routes without knowing
the external gateway prefix. Service-token exchanges and the Flipt authorization
callback keep their existing targets.

Run the automated facade checks before rehearsing against the dev stack:

```sh
bazel test //services/tadoku-api/...
bazel test //services/tadoku-api/transport/http:http_test --test_filter='^$' --test_arg=-test.bench=BenchmarkFacade --test_arg=-test.benchtime=1s --test_output=all
```

The benchmark compares direct and proxied loopback HTTP requests. It does not
measure deployed gateway latency or replace the full application contract run.

Frontend pods use Tilt `live_update` with scoped sync paths and polling file watchers for Next.js.
Routine edits under a frontend app or `frontend/packages/ui` should sync into the running pod and hot-reload without an image rebuild; package file changes run `pnpm -r install` inside the container.

Registry cache is intentionally skipped for now: the dev stack already pulls from the configured per-environment registry, and there is no committed evidence that image pulls are a bottleneck.

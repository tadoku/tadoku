# Tadoku Monorepo

[![Documentation](https://img.shields.io/badge/docs-online-6969FF.svg)](https://tadoku.github.io/tadoku/)
![Build Bazel](https://github.com/tadoku/tadoku/actions/workflows/build-bazel.yaml/badge.svg)
![Build Frontend Web](https://github.com/tadoku/tadoku/actions/workflows/build-frontend-web.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)

Tadoku had a significant rewrite and the documentation hasn't been updated yet to reflect these changes.
The documentation for this repository can be found at https://tadoku.github.io/tadoku/.

## Dev Environment

Use `k8s/dev/` through the root `Tiltfile` for both shared and local clusters.
Copy `tilt_config.json.example` to `tilt_config.json` for machine-specific hostnames, registries, and contexts.

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

Frontend pods use Tilt `live_update` with scoped sync paths and polling file watchers for Next.js.
Routine edits under a frontend app or `frontend/packages/ui` should sync into the running pod and hot-reload without an image rebuild; package file changes run `pnpm -r install` inside the container.

For clusters that pull from a private registry, set `TADOKU_DEV_REGISTRY_HOST` (or the `registry` value in `tilt_config.json`), `TADOKU_DEV_REGISTRY_USERNAME`, and `TADOKU_DEV_REGISTRY_PASSWORD` before `make dev-up` so Tilt creates the `tadoku-dev-registry-pull` secrets; see `docs/docs/local-environment.md` for details.

Registry cache is intentionally skipped for now: the dev stack already pulls from the configured per-environment registry, and there is no committed evidence that image pulls are a bottleneck.

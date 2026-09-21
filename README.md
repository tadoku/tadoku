# Tadoku Monorepo

[![Documentation](https://img.shields.io/badge/docs-online-6969FF.svg)](https://tadoku.github.io/tadoku/)
![Build Bazel](https://github.com/tadoku/tadoku/actions/workflows/build-bazel.yaml/badge.svg)
![Build Frontend Web](https://github.com/tadoku/tadoku/actions/workflows/build-frontend-web.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)

Tadoku had a significant rewrite and the documentation hasn't been updated yet to reflect these changes.
The documentation for this repository can be found at https://tadoku.github.io/tadoku/.

## Dev Environment

Use **DevCLI** for webv2 and native Tadoku API development on
[https://tadoku.dev.lab](https://tadoku.dev.lab). See
[the development runbook](.dev/README.md) for installation, shared-stack
prerequisites, branch databases, routing and cleanup.

```sh
dev doctor
make dev-seed  # shared synthetic users and base fixtures
DEV_OWNER=anton make dev-up
# In another terminal, same checkout:
dev url --owner anton '/'
dev status --owner anton
dev logs --owner anton tadoku-api
dev down --owner anton
```

Keep the development loop running: frontend edits use the pnpm Next.js dev
server/HMR; backend edits rebuild the affected Bazel binary in a long-lived pod.
The printed link selects your environment without copying branch strings.

One shared Postgres pod holds the base/auth databases and a unique application
database per branch/owner. Migration and seed Jobs run explicitly before API
startup; branch databases survive `dev down`. Kratos and Keto stay shared.
The existing seeder provides `dev@tadoku.app` and `reader@tadoku.app` with the
development fixture password `tadoku`, unless overridden outside Git.

Legacy Tilt files remain for base infrastructure; do not run Tilt concurrently
with DevCLI because it can replace canonical routing. Full base GitOps adoption
is separate work. `make dev-reset` is a destructive shared-database reset, not
normal branch cleanup.

# Tadoku Monorepo

[![Documentation](https://img.shields.io/badge/docs-online-6969FF.svg)](https://tadoku.github.io/tadoku/)
![Build Bazel](https://github.com/tadoku/tadoku/actions/workflows/build-bazel.yaml/badge.svg)
![Build Frontend webv2](https://github.com/tadoku/tadoku/actions/workflows/build-frontend-webv2.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/tadoku/tadoku)](https://goreportcard.com/report/github.com/tadoku/tadoku)

Developer documentation lives in [`docs/docs`](docs/docs) and is published at https://tadoku.github.io/tadoku/.

## Dev Environment

New here? Start with the [agent verification guide](.agents/skills/verify-tadoku/SKILL.md)
and [feature map](.agents/skills/verify-tadoku/references/features/README.md).
They explain how to find a user journey, select your branch, prove a change in
the browser, and clean up. Agents can use `$verify-tadoku` when repository skill
discovery is supported, or read the linked guide directly; no global skill is required.

Use **DevCLI v0.4.0+** for webv2, auth, admin and native Tadoku API development on
[https://tadoku.dev.lab](https://tadoku.dev.lab). See
[the development runbook](.dev/README.md) for installation, shared-stack
prerequisites, branch databases, routing and cleanup.

```sh
dev doctor
make dev-seed  # shared synthetic users and base fixtures
DEV_OWNER=anton make dev-up
# In another terminal, same checkout:
dev url --owner anton '/'
dev url --owner anton --host account.tadoku.dev.lab '/login'
dev url --owner anton --host admin.tadoku.dev.lab '/'
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

The fresh **development-only** base is defined in
[`k8s/dev/base/`](k8s/dev/base/README.md) for Argo CD, with automatic migrations
and existing GHCR images tracked by development Image Updater. It is active on
`homelab-dev` in the `tdk-dev-*` namespaces. Token-reflector is base-only;
Paper styleguide overlays are deferred and its local pnpm workflow remains.
`make dev-reset` is disabled; any reset needs an explicitly approved, scoped runbook.

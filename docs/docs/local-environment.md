---
sidebar_position: 3
title: Development Environment
---

# Development Environment

Use DevCLI for live edits to webv2, auth, admin and the native Tadoku API.
Argo CD keeps the shared base running on **homelab-dev** even when no developer
has a CLI loop running. This is development infrastructure, not the production
deployment.

The repository's [DevCLI runbook](https://github.com/tadoku/tadoku/blob/main/.dev/README.md)
owns the full workflow. The [base runbook](https://github.com/tadoku/tadoku/blob/main/k8s/dev/base/README.md)
owns operator bootstrap, credentials, automatic migrations and GitOps recovery.

## Prerequisites

- Git access to Tadoku and the private `antonve/dev-cli` repository.
- Go for installing DevCLI, Bazelisk/Bazel using the repository's pinned version,
  and kubectl with authorized access to the `homelab-dev` context.
- Node and pnpm for frontend tools; Docker for local image/build checks.
- Lab network/DNS access and trust in the Lab CA, including the configured
  development registry at `registry.dev.lab`.

The existing cluster supplies Postgres, shared auth providers and routing.
There is no local Kubernetes cluster or Helm bootstrap to run. Obtain kube
access from the operator; never commit kubeconfigs, credentials or private keys.
The non-secret configuration is committed in `.dev/config.yaml`.

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@latest
dev version  # v0.4.0 or newer; put the Go install directory on PATH
git fetch origin main
dev doctor
```

If Git authentication is SSH-only, use the installation variant in the DevCLI
runbook. Doctor checks prerequisites without deploying; it is not an end-to-end
routing or login test.

## Start a branch

Create a feature branch and edit a service before starting. DevCLI asks Bazel
which deployables changed relative to the merge base with `origin/main`, including
uncommitted edits. It creates only affected overlays; the other services keep
using base. Unknown paths conservatively select all deployables.

```sh
make dev-seed  # shared synthetic identities and base fixtures; safe to rerun
dev up --owner alice --task migrate --task seed
```

Leave this terminal running. Frontend sources sync into a pnpm-managed Next.js
dev server with HMR. Backend edits rebuild the affected Bazel binary and restart
it in the existing pod; failed compilation keeps the last working process.
Development images are built on demand and pushed to the configured registry.
The base instead uses CI-published GHCR images tracked by development Image Updater.

Owner is a stable developer/worktree label, not authentication. Use a distinct
owner for simultaneous checkouts. Restart the loop when you need to add a service.
Frontend-only work can omit the migration/seed tasks. `make dev-up` wraps the
command with those tasks; use `DEV_OWNER=alice` consistently for Make wrappers.

## Open and inspect your environment

In another terminal, from the same checkout:

```sh
dev url --owner alice '/'
dev url --owner alice --host account.tadoku.dev.lab '/login'
dev url --owner alice --host admin.tadoku.dev.lab '/'
dev status --owner alice
dev logs --owner alice tadoku-api
```

Open the printed link to select the branch automatically. Envoy sets a host-only
cookie; no branch string needs copying into an application. Selection is per
hostname, so open the main-host link too when testing an API overlay from admin.
Separate browser profiles have independent selections; tabs in one profile share
the cookie. Clear a hostname with `dev url --owner alice --clear '/'` (add
`--host` for account or admin).

Without a selection, browsers use the base at `https://tadoku.dev.lab`,
`https://account.tadoku.dev.lab` and `https://admin.tadoku.dev.lab`. Each service
independently falls back to base when no healthy selected overlay exists. Check
`X-Dev-Backend` to confirm the actual upstream; cold-start routing convergence
can temporarily serve base even after the CLI prints its link.

Token-reflector is base-only. Paper styleguide live overlays are deferred;
its local workflow is still `cd frontend && pnpm paper-styleguide`.

## Migrations and seed data

Argo CD automatically runs base migrations during full syncs before dependent
workloads start. Do not use selective sync for releases because it skips hooks.
Base seed data is explicit: `make dev-seed` creates the synthetic administrator
`dev@tadoku.app` and reader `reader@tadoku.app`, with fixture password `tadoku`
unless overridden outside Git. It marks identities it owns and refuses to take
over unmarked accounts. Kratos and Keto remain shared.

One operator-managed Postgres pod holds separate application databases for each
owner/branch. `dev up --task migrate --task seed` provisions and prepares the
selected branch database before API startup. Tasks can also be rerun explicitly:

```sh
dev task --owner alice migrate
dev task --owner alice seed
```

These tasks do not reset databases. Branch workers and cache prefixes isolate
leaderboards on shared Valkey. This is cooperative development isolation, not
hostile multi-tenancy. Deploy related services together when a test requires
consistent data across services otherwise using different base/branch data.

## Cleanup and troubleshooting

```sh
dev down --owner alice
```

This stops the local loop and removes only that owner/branch's overlay resources;
it does not stop the Argo base. Abandoned overlays can be removed through the
CLI's explicit TTL cleanup. Branch databases are retained after either cleanup.
`make dev-reset` fails closed: database deletion needs separate approval and an
exactly scoped runbook, never namespace-wide deletion.

Use status and service-filtered logs for sync/build errors. Check the selected
hostname and backend headers before diagnosing stale content. Successful backend
restarts may briefly return 503; compile failures preserve the old process.
With the loop running, replacement pods receive the current source again.
Without it, a replacement starts from its image, not a previous pod's edits.

For a shared-base failure, inspect the Argo application and follow the base
runbook; do not restart or delete shared databases as a troubleshooting shortcut.

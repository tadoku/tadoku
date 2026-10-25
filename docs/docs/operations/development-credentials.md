---
title: Development credentials
description: How the development base generates database and provider credentials, references them in workloads and handles rotation.
---

# Development credentials

Read this when bootstrapping or rotating credentials for the development base.
[Development base](development-base.md) covers activation and data safety.

There are no plaintext Secret manifests or private keys in `k8s/dev/base/`.

- The Postgres operator generates `tadoku_owner` and `tadoku`
  credentials in `tdk-dev-data`. The owner Secret is named
  `tadoku-owner.tadoku-dev-db.credentials.postgresql.acid.zalan.do`; the
  operator replaces underscores in Secret names. The base Tadoku migrations
  use this owner Secret. Tenant lifecycle Jobs use it for tenant and override
  writes. API, worker and SQL fixture writes keep the `tadoku` credentials.
- The former Keto database role and its generated credentials are retained;
  no bootstrap step copies them and no deployed service consumes them. Their
  deletion follows the retained-state rules in
  [Development base](development-base.md#data-safety).
- The Postgres administrator Secret stays in `tdk-dev-data`. The ownership
  hooks and short-lived branch database creation Job use it. Branch migration
  tasks use `tadoku_owner`; ordinary tenant tasks do not use administrator
  credentials. API, worker and fixture seeding use `tadoku`. This
  is cooperative isolation, not hostile multi-tenancy.
- `scripts/dev/bootstrap-gitops-secrets.sh` copies only the required
  credentials, including the owner credential for branch migrations, into
  consumer namespaces. It targets only `homelab-talos-dev` at
  `https://omni.lab:8100`. It generates the development-only Oathkeeper JWKS
  and Oathkeeper authorization token once, and
  preserves existing keys on rerun. It needs `kubectl`, `jq` and `node`.
- The script fails closed on another API server, on namespaces not labeled as
  part of the development base, and on Secrets with unexpected ownership. It
  never reads credentials from outside the development GitOps base.
- Run it only with explicit credential or bootstrap approval.
- Keep generated keys only in development Kubernetes Secrets or an approved
  encrypted backup, never in Git or a worktree. Commit only the public Lab CA
  and Secret references, never private keys, kubeconfigs, tokens or plaintext
  Secrets.

When operator credentials rotate, rerun the bootstrap script, then perform an
explicitly approved restart of the consumers. Kubernetes does not refresh
environment variables in running processes.

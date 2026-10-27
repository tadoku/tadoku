---
title: Operations
description: The Tadoku runbooks for the development base, failed migrations and identity incident recovery, plus the content and backup policies that apply to operations.
sidebar_position: 1
---

# Operations

Read this when you need a runbook for the development base or a manual data repair.

## Runbooks

| Runbook | Use it to |
| --- | --- |
| [Development base](./development-base.md) | Operate the development-only base in `k8s/dev/base/`: ownership, automatic migrations, credential bootstrap and verification |
| [Database migration recovery](./migration-recovery.md) | Contain and repair a failed Tadoku API migration that left `schema_migrations` dirty |
| [Identity incident recovery](./identity-recovery.md) | Decide whether to revert identity changes or restore identity data during a production incident, and verify database recovery |

## CMS-managed content

Public page copy is edited in the admin CMS only; see
[Contributing workflow](../develop/contributing.md#cms-managed-content).

## Backups

The production database runs on PlanetScale, and Tadoku has no separate backup
job. Its backup schedule, retention and point-in-time recovery coverage come
from the provider account and are not recorded here. Verify them there before
relying on a restore; see
[Database recovery](./identity-recovery.md#database-recovery).

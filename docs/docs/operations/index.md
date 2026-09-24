---
title: Operations
description: The Tadoku runbooks for the development base, failed migrations and account deletion, plus the content and backup policies that apply to operations.
sidebar_position: 1
---

# Operations

Read this when you need a runbook for the development base or a manual data repair.

## Runbooks

| Runbook | Use it to |
| --- | --- |
| [Development base](./development-base.md) | Operate the development-only base in `k8s/dev/base/`: ownership, automatic migrations, credential bootstrap and verification |
| [Database migration recovery](./migration-recovery.md) | Contain and repair a failed Tadoku API migration that left `schema_migrations` dirty |
| [Account deletion](./account-deletion.md) | Delete a user's data manually and leave an anonymised tombstone row |

## CMS-managed content

Public page copy is edited in the admin CMS only; see
[Contributing workflow](../develop/contributing.md#cms-managed-content).

## Backups

PlanetScale takes the production database backups and point-in-time recovery
points; there is no separate backup job.

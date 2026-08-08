# Ory stack upgrade plan

<!-- markdownlint-disable MD013 -->

Status: planning

Production authority: `lke20948-ctx` and Argo CD

Observed GitOps revision: `a9100d9e`

Target: Ory chart `0.63.0`, applications `v26.2.0`

Rehearsal policy: one successful timed rehearsal on `homelab-dev`
Soak policy: no business soak; use immediate verification gates

## Overview

Upgrade Keto and Oathkeeper independently, then upgrade Kratos with a controlled big-bang cutover.

Kratos cannot use a normal rolling deployment. Its `v26.2.0` PostgreSQL migration adds and backfills `identity_id` on `identity_credential_identifiers` and `session_devices`, then makes the field `not null`. Kratos `v0.13.0` does not populate that field, so an old pod must not write after the migration begins.

The selected cutover is:

1. Prepare and rehearse everything outside the production window.
2. Hold Argo reconciliation and stop the Kratos API and courier.
3. Take a final Kratos database copy.
4. Deploy the standalone Kratos migration.
5. Base and deploy the Kratos runtime change on the migrated `main` branch immediately.
6. Run critical verification and restore traffic.
7. Introduce Sealed Secrets and rotate all secrets afterward as separate work.

Target downtime is at most five minutes from the last `v0.13.0` writer stopping to healthy `v26.2.0`. The production window must use the duration measured by the single rehearsal if it exceeds five minutes.

## Current production state

| Service | Chart | Application | Database behavior | Namespace |
| --- | --- | --- | --- | --- |
| Kratos | `0.32.0` | `v0.13.0` | PostgreSQL, chart automigration enabled | `tdk-prod-kratos` |
| Keto | `0.60.1` | `v25.4.0` | PostgreSQL, init-container automigration enabled | `tdk-prod-keto` |
| Oathkeeper | `0.60.1` | `v25.4.0` | Stateless | `tdk-prod-oathkeeper` |

All three Argo Applications are healthy and synced with automated prune and self-heal enabled. There is no sync window. Merging a values change into Git can therefore become a production deployment.

Current Kratos deployment facts:

- Deployment strategy: `RollingUpdate`
- `maxSurge`: `30%`
- `maxUnavailable`: `0`
- API replicas: 1
- Courier replicas: 1
- Kratos database size: approximately 133 MiB
- Identities: 4,320
- Credential identifiers: 4,320
- Sessions: 6,113
- Session devices: 6,113
- Last three production backup durations: 35, 36, and 61 seconds

The two v26 PostgreSQL backfills process up to 10,000 rows per batch. Each affected production table currently fits into one batch.

## Migration compatibility findings

| Component | Finding | Operational consequence |
| --- | --- | --- |
| Kratos `v0.13.0 → v26.2.0` | Not backward-compatible for writes | Stop every old API and courier pod before migration. Do not use a rolling update. |
| Keto `v25.4.0 → v26.2.0` | No SQL/Go schema migration was added | Run `migrate sql status`; if clean, deploy the runtime without a migration PR. |
| Oathkeeper | No database | Runtime/configuration rollback remains available through Argo. |

The Kratos finding is a source-level conclusion that must be reproduced during rehearsal:

- v26.2.0 adds nullable `identity_id` columns.
- It backfills credential identifiers and session devices.
- It enforces `not null` and foreign keys.
- The v0.13.0 credential-identifier and session-device models do not write `identity_id`.

## Benefits

- Eliminates the unsafe period where old and new Kratos writers share the migrated database.
- Moves image download, manifest rendering, contract testing, and rollback preparation outside downtime.
- Keeps the vendor migration and dependent runtime in separate changes.
- Makes downtime measurable: final copy, migration, readiness, and critical smoke tests.
- Keeps Keto, Oathkeeper, Sealed Secrets, and secret rotation outside the Kratos blast radius.

## Risks and controls

| Risk | Control | Stop condition |
| --- | --- | --- |
| Argo auto-sync deploys unexpectedly | Establish and test a sync hold before version changes | A test revision can reconcile without an explicit operator release |
| Rolling Kratos overlaps incompatible writers | Disable automigration, use `Recreate`, and independently confirm API/courier replicas are zero | Any `v0.13.0` container is still running when migration starts |
| Migration or new runtime fails | Take a restorable final copy and keep complete restore commands ready | The final copy is incomplete or has not passed restore rehearsal |
| Runtime PR preparation extends downtime | Prepare an unsubmitted patch and CI evidence; have a reviewer waiting; after migration deployment, base the PR on updated `main` | The target patch cannot be reviewed and merged immediately after migration |
| Authenticated traffic fails during the window | Return a deliberate maintenance response and keep the window short | Traffic is not in maintenance mode before Kratos reaches zero replicas |
| Secret work obscures an Ory regression | Keep all secret values unchanged until Ory passes immediate verification | A credential or signing-key change enters an Ory migration/runtime PR |
| Final backup increases downtime | Use the measured 35–61 second logical copy for zero-RPO rollback | The copy exceeds the rehearsed window or cannot be restored |

Taking the final copy after writers stop gives a zero recovery-point objective but adds roughly one minute. Taking it immediately before the freeze reduces downtime but can lose subsequent writes on rollback. The default is the zero-RPO option.

## Measurable success criteria

- [ ] One production-copy rehearsal completes the full Kratos sequence successfully.
- [ ] Every segment is timed: writer shutdown, final copy, migration, runtime PR/sync, readiness, critical smoke tests, and restoration.
- [ ] The production budget is at most five minutes, or is replaced with the slower measured rehearsal duration.
- [ ] Argo cannot reconcile any Ory Application during the hold without operator action.
- [ ] Kratos API and courier replicas are both zero before the migration Job starts.
- [ ] Chart automigration is disabled and the target Kratos Deployment renders with `strategy.type: Recreate`.
- [ ] The migration is a standalone commit, PR, and deployment.
- [ ] The dependent runtime change is based on `main` only after the migration has deployed.
- [ ] Existing sessions still resolve through `/sessions/whoami` after cutover.
- [ ] New login, registration, credential-identifier creation, session creation, and session-device creation succeed.
- [ ] Recovery, verification, settings, logout, courier delivery, redirects, cookies, and CORS pass.
- [ ] Keto reports no schema delta and its permission contracts pass on `v26.2.0`.
- [ ] All five Oathkeeper rules and both user/service token contracts pass.
- [ ] No secret value appears in tickets, PR text, logs, rendered artifacts, or this plan.
- [ ] Sealed Secrets adoption and secret rotation begin only after immediate Ory verification passes.

## Parallel work map

| Lane | Subagent scope | Dependency |
| --- | --- | --- |
| A | Argo deployment hold and render-diff checks | Must pass before any production mutation |
| B | Kratos production copy, restore, migration timing, and rollback timing | New isolated `homelab-dev` database |
| C | Keto production copy and migration-status verification | Separate new `homelab-dev` database |
| D | Kratos browser/session contract suite | Can run with B after the restore is healthy |
| E | Keto and Oathkeeper permission/proxy/token suites | Independent of Kratos migration work |
| F | Frontend Ory adapter modernization | Runs in parallel with backend work |
| G | Go Kratos/Keto adapter modernization | Runs in parallel with frontend work |

Production database changes and runtime rollouts are sequential and must not be delegated as concurrent production operations.

## Phase 1 — Establish production control

- [ ] Add a tested Argo sync window or temporary manual-sync procedure for `tadoku-kratos`, `tadoku-keto`, and `tadoku-oathkeeper`.
- [ ] Prove that an operator can hold and release reconciliation.
- [ ] Pin the intended Git revision, chart versions, and image digests.
- [ ] Add manifest render/diff checks for all three multi-source Applications.
- [ ] Record the current images, replica counts, service selectors, health checks, and database migration status.
- [ ] Keep every current credential and signing key unchanged.

Gate: Argo reconciliation is held predictably, a test revision cannot deploy automatically, and the captured baseline contains no secret values.

## Phase 2 — Create the rehearsal databases

- [ ] Take production copies of the Kratos and Keto logical databases.
- [ ] Create uniquely named, new Kratos and Keto databases on `homelab-dev`.
- [ ] Restore each production copy into its corresponding new database.
- [ ] Point only disposable rehearsal workloads at the new database DSNs.
- [ ] Verify migration ledgers, row counts, identity/session records, and Keto tuples.
- [ ] Verify current Kratos and Keto binaries are healthy against the restored copies.
- [ ] Record the copy and restore durations.

Gate: both new `homelab-dev` databases reproduce the expected production counts and pass current-version health checks.

## Phase 3 — Build the contract harness

Independent subagents may implement these suites in parallel.

- [ ] Kratos: existing session, login, registration, logout, recovery, verification, settings, and courier behavior.
- [ ] Kratos: cookie domain, SameSite, CORS, base paths, return URLs, and redirect behavior.
- [ ] Keto: known allow/deny cases, tuple create/list/delete, batch result association, pagination, OPL namespace, and PostSync seed.
- [ ] Oathkeeper: anonymous access, cookie session, HTML redirect, JSON `401`, four API path strips, user JWT, and service-token exchange.
- [ ] Capture representative JWT claim fixtures without including signing material.
- [ ] Run read-only checks against production and destructive checks only against `homelab-dev`.

Gate: critical contracts pass on the current production versions and against the restored rehearsal environment.

## Phase 4 — Modernize application clients

Frontend and backend work are independent and may run in parallel.

- [ ] Put frontend Kratos calls behind a local adapter.
- [ ] Replace legacy `V0alpha2Api` usage with the current component SDK.
- [ ] Keep UI-node rendering exhaustive and fail clearly on unknown node types.
- [ ] Preserve narrow Go interfaces and update Kratos/Keto generated clients behind them.
- [ ] Add/update frontend and backend tests for every touched adapter behavior.
- [ ] Deploy client changes against the current production Ory versions before changing servers.
- [ ] Run frontend typecheck, lint, and build plus the relevant Bazel builds/tests.

Gate: modern clients pass the contract harness against Kratos `v0.13.0` and Keto/Oathkeeper `v25.4.0`.

## Phase 5 — Run the single timed rehearsal

- [ ] Disable automigration in the rehearsal chart values.
- [ ] Render Kratos `v26.2.0` with `strategy.type: Recreate`.
- [ ] Pin the migration Job and runtime to exact image digests.
- [ ] Stop the rehearsal Kratos API and courier and confirm both have zero replicas.
- [ ] Take the final rehearsal database copy and start the downtime timer.
- [ ] Run the standalone `v26.2.0` migration Job.
- [ ] Deploy the `v26.2.0` runtime immediately.
- [ ] Run the critical verification suite and stop the downtime timer.
- [ ] Complete the wider verification suite.
- [ ] Restore the pre-migration copy into a new database and prove rollback health with `v0.13.0`.
- [ ] Record every segment and publish the measured production window.
- [ ] Confirm Keto reports no pending schema migration on its production copy.

Gate: the single rehearsal and restoration both pass, every segment has a measured duration, and the production window uses the measured result.

## Phase 6 — Upgrade Keto and Oathkeeper

- [ ] Deploy values that disable Keto runtime automigration while keeping the current runtime unchanged.
- [ ] Run `migrate sql status`; stop if it reports an unexpected pending migration.
- [ ] Upgrade Keto to chart `0.63.0` and application `v26.2.0`.
- [ ] Verify permission checks, tuple operations, pagination, OPL, and the deterministic admin seed.
- [ ] Upgrade Oathkeeper to chart `0.63.0` and application `v26.2.0`.
- [ ] Retain the active signing material.
- [ ] Verify all five access rules, forwarded-header behavior, redirects, user JWTs, and service tokens.

Gate: Keto and Oathkeeper pass their complete immediate verification suites with no unexplained Argo drift.

## Phase 7 — Execute the production Kratos big bang

Before the window:

- [ ] Disable Kratos chart automigration while retaining `v0.13.0`.
- [ ] Prepare and approve the standalone migration PR.
- [ ] Prepare an unsubmitted runtime patch, expected render, CI evidence, and a waiting reviewer.
- [ ] Pre-pull the exact `v26.2.0` image digest on every eligible LKE node.
- [ ] Prepare exact scale-down, migration, runtime-sync, smoke-test, and restore commands.
- [ ] Put Kratos-dependent traffic into maintenance mode.

During the window:

- [ ] Hold Argo reconciliation.
- [ ] Scale the Kratos Deployment and courier StatefulSet to zero.
- [ ] Confirm no `v0.13.0` API, courier, init, or migration container is running.
- [ ] Take the final logical database copy and confirm completion.
- [ ] Deploy the standalone migration PR/Job and verify a clean migration ledger.
- [ ] Base the runtime PR on the now-migrated `main` branch.
- [ ] Merge and sync chart `0.63.0` with Kratos `v26.2.0`, `Recreate`, and automigration disabled.
- [ ] Verify readiness, existing `/sessions/whoami`, new login, new registration, credential identifiers, sessions, session devices, and courier delivery.
- [ ] Restore normal traffic.
- [ ] Complete the remaining Kratos and Oathkeeper session checks.
- [ ] Re-enable normal Argo automation after cleanup review.

Gate: downtime is within the rehearsed budget and every critical and extended Kratos contract passes. If the critical gate fails, restore the complete pre-migration database before starting `v0.13.0` again.

## Phase 8 — Adopt Sealed Secrets and rotate secrets

This phase starts only after the immediate Ory verification gate. It is not part of the Ory migration.

- [ ] Deploy and verify the Sealed Secrets control plane independently.
- [ ] Move the current database, SMTP, backup, webhook, and Oathkeeper signing values into sealed resources without changing their effective values.
- [ ] Verify workload health, courier delivery, backups, authorization, sessions, and JWT verification.
- [ ] Remove plaintext secret values from the active Git tree.
- [ ] Rotate database credentials and verify every consumer.
- [ ] Rotate SMTP credentials and verify courier delivery.
- [ ] Rotate backup and webhook credentials and run a backup.
- [ ] Rotate Oathkeeper signing material with verification overlap where required.

Gate: every live secret is supplied through Sealed Secrets, all secret categories have been rotated and verified, and the active Git tree contains no plaintext secret material.

## Atomic PR sequence

| PR | Change | Execution |
| --- | --- | --- |
| A | Add Argo deployment gates and render checks | Parallel batch 1 |
| B | Add the Ory contract harness | Parallel batch 1 |
| C | Modernize the frontend Ory adapter | Parallel batch 2 |
| D | Modernize the Go Ory adapters | Parallel batch 2 |
| E | Disable runtime automigration without changing versions | Sequential |
| F | Upgrade Keto; migration status must be clean | Sequential |
| G | Upgrade Oathkeeper | Sequential |
| H | Deploy the standalone Kratos migration while old writers are stopped | Maintenance window |
| I | Upgrade Kratos from updated `main` immediately after PR H deploys | Maintenance window |
| J | Remove temporary Jobs and restore normal Argo flow | Sequential |
| K | Adopt Sealed Secrets with unchanged values | Post-migration |
| L | Rotate and verify all secret categories | Post-migration |

PR H and PR I must remain separate. PR I must be based on `main` only after PR H has landed and deployed.

## Rollback

### Before the Kratos migration

Release the Argo hold and restore the previous desired state. No database rollback is required.

### After migration, before Kratos verification passes

1. Keep traffic in maintenance mode.
2. Scale the failed `v26.2.0` runtime to zero.
3. Restore the complete pre-migration Kratos logical database.
4. Reconcile the `v0.13.0` runtime revision.
5. Verify login and `/sessions/whoami` before reopening traffic.

Do not run an unrehearsed down migration or restore selected tables in place.

Keto has no schema delta on this exact path, and Oathkeeper is stateless. Their runtime/chart revisions can be reverted independently through Argo while retaining Keto tuple data and Oathkeeper signing continuity.

## Follow-up improvements

- Add a deliberate authentication maintenance page and API `503` contract so cutovers fail clearly.
- Keep a reusable preflight that checks image availability, replica zero, migration status, timers, smoke tests, and rollback readiness.
- Schedule recurring Kratos and Keto logical-restore drills.
- Evaluate two Kratos replicas and a PodDisruptionBudget after the upgrade.
- Add CI that compares Ory migration trees and flags old-binary incompatibility before future version bumps.

## References

- [Ory Helm chart index](https://k8s.ory.sh/helm/charts/index.yaml)
- [Ory Kubernetes chart v0.63.0](https://github.com/ory/k8s/releases/tag/v0.63.0)
- [Kratos chart v0.63.0 values](https://github.com/ory/k8s/blob/v0.63.0/helm/charts/kratos/values.yaml)
- [Kratos v26.2.0 nullable `identity_id` columns](https://github.com/ory/kratos/blob/v26.2.0/persistence/sql/migrations/sql/20251104000000000000_identifiers_devices_identity_id.up.sql)
- [Kratos v26.2.0 batched PostgreSQL backfill](https://github.com/ory/kratos/blob/v26.2.0/persistence/sql/migrations/go/20251105000000000000_identity_id_not_null_fks.go)
- [Kratos v26.2.0 `not null` and foreign-key enforcement](https://github.com/ory/kratos/blob/v26.2.0/persistence/sql/migrations/sql/20251105000000000003_identity_id_not_null_fks.postgres.up.sql)
- [Kratos v0.13.0 credential-identifier model](https://github.com/ory/kratos/blob/v0.13.0/identity/credentials.go#L116-L132)
- [Kratos v0.13.0 session-device model](https://github.com/ory/kratos/blob/v0.13.0/session/session.go#L45-L70)
- [Keto v25.4.0 to v26.2.0 comparison](https://github.com/ory/keto/compare/v25.4.0...v26.2.0)
- [Kratos v26.2.0 release](https://github.com/ory/kratos/releases/tag/v26.2.0)
- [Keto v26.2.0 release](https://github.com/ory/keto/releases/tag/v26.2.0)
- [Oathkeeper v26.2.0 release](https://github.com/ory/oathkeeper/releases/tag/v26.2.0)

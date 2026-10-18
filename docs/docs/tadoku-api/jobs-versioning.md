---
title: Migrate a job contract from v1 to v2
description: Introducing a new job contract version, deploying consumers before producers, draining and retiring the old version, and rollback compatibility.
sidebar_position: 5.2
---

# Migrate a job contract from v1 to v2

Read this when changing the payload shape, validation or meaning of a stored
job. [Jobs and worker](./jobs.md) covers defining and registering a job.

A versioned name and its payload form a protocol between independently deployed
API and worker images, and stored rows outlive both. Use a new version when the
payload shape, validation or meaning changes. Run both versions while v1 drains;
never rewrite stored rows. These facts drive the sequence:

- A worker claims only names it registers. Rows with other names stay `pending`
  and appear in the unsupported gauges.
- Decoding is strict. A v1 handler fails a payload with an extra field as
  `invalid_payload`.
- Replay copies the failed row's name and payload unchanged, and only for a
  name the replaying binary registers.
- Failed rows stay in `jobs` indefinitely, in every tenant.

## Define and register both versions

Add a distinct v2 value, name and validation, and register its handler beside
v1 in `registrations`. Keep the v1 payload and handler intact. Define whether
v1 and v2 effects can overlap safely and how idempotency works across them.
Verify old persisted v1 payloads and new v2 payloads against their handlers.

**Gate:** both versions can be processed, replay behavior is understood, and
supported producer and worker rollback versions have been identified.

## Deploy consumers, then switch producers

Release the worker that registers v1 and v2 before any producer emits v2, and
verify the running revision processes v2 through a controlled workflow. Then
release the producers that emit v2. Producer replicas still being replaced may
write v1 meanwhile, which is safe while v1 is registered. Do not enqueue both
versions for one business action.

**Gate:** all intended producers emit v2, the worker processes both versions,
and the unsupported pending and running gauges are zero.

## Drain v1 and decide on its failures

Count v1 rows in every tenant with the
[all-tenants query](./jobs-operations.md#inspect-and-replay-retained-work);
branch tenants share the queue. Pending rows include delayed retries, and
running rows hold leases; let the worker drain them. For each failed v1 row,
either repair the cause and replay it while v1 is registered, or record why it
needs no replay. Replays of v1 are v1 rows too, and a successful replay leaves
its failed original in place.

**Gate:** no pending or running v1 rows remain in any tenant, and every failed
v1 row has a recorded replay or no-replay decision.

## Retire v1

Remove the v1 registration only when all of these conditions hold:

- No current producer can emit v1.
- No supported producer rollback can emit v1.
- Every supported worker rollback can execute every version that may be queued.
- The previous gates passed.

Delete the v1 `handle` entry and its `registrations` parameter; remove the
payload type once nothing references it. Failed v1 rows then remain as history:

- They can no longer be replayed; the replay command answers
  `job is not failed or supported`. Replaying one later requires releasing a
  worker that registers v1 again. Never redirect it to v2.
- They count in `tadoku_worker_unsupported_failed_jobs`, and the worker logs
  `unsupported job types remain outstanding` at startup while any remain. Use
  the unsupported pending and running gauges to detect stranded new work.
- The queue has no retirement flag, archive or conversion command. Any new
  archive or conversion workflow requires its own review, preserves history and
  replay lineage, and ships any SQL as a standalone migration.

## Rollback and optional fields

After producers start writing v2, reverting every consumer to v1 strands v2
rows even if producers have already reverted. Keep a compatible consumer for
every version that remains queued or may still be published. A producer rollback
to v1 is safe only while v1 support remains. These are compatibility conditions,
not permission to perform a rollback.

Strict decoding rejects unknown fields. Adding an apparently optional JSON
field can therefore break an older worker. Prefer a new version. A same-version
extension requires readers that accept both forms to be deployed everywhere
first, with old/new fixtures and the supported rollback direction verified.
Do not silently relax decoding for the entire catalogue.

If draining both versions is insufficient, design a separate reviewed data
conversion with deterministic mapping, running-claim handling, idempotency,
audit lineage and reconciliation. Never bulk relabel stored v1 rows as v2.
A SQL data conversion remains a standalone migration PR and deployment.

---
title: Database migration recovery
description: How to contain and repair a failed Tadoku API migration that left schema_migrations dirty, using the inspect and force recovery commands.
---

# Database migration recovery

This runbook applies when Tadoku API's Argo CD migration Job fails
and `schema_migrations.dirty` is `true`.

The normal `/migrate` command deliberately supports only forward `up`
migrations. The separate `/migrate-recovery` command supports:

- `inspect`, which reads the current migration version and dirty state;
- `force`, which changes migration metadata without running SQL.

It does not support `down`, `drop`, or arbitrary migration steps.

## Safety rules

1. Never use `force` to make an error disappear. It changes only
   `schema_migrations`; it does not repair tables or data.
2. Stop automated retries before inspecting or repairing the database.
3. Ensure exactly one migration or recovery process can access the affected
   database.
4. Record the failed image digest, migration version, Job logs, Git revision,
   database state, and recovery commands in the incident.
5. Take and verify a database backup or PITR recovery point before manual SQL
   or metadata repair.
6. Have a human review the physical schema assessment, repair SQL, and
   selected target version.
7. Prefer a forward fix. Do not run destructive down migrations during an
   incident.

## 1. Contain the rollout

1. Terminate the affected Argo CD sync operation.
2. Disable automated sync and image updates for the affected Application.
3. Confirm that the failed migration Job did not deploy the new API image.
4. Confirm that the previous API Deployment is available and passes basic
   health checks.
5. Confirm that no migration or recovery Pod is currently running.

Do not start another Argo sync while recovery is in progress.

## 2. Preserve evidence and backup

PlanetScale takes the production database backups and provides the point-in-time
recovery (PITR) points referenced in this runbook. There is no separate backup job.

Capture:

- the failed Job YAML and Pod logs;
- the Application operation result;
- the migrations image digest;
- the migration filenames embedded in that image;
- the current database backup status and PITR recovery point.

Create an incident-specific backup before modifying the database. Verify that
the backup can be listed or restored in an isolated environment.

## 3. Inspect migration metadata

The `/migrate-recovery` binary is included in `tadoku-api-migrations` images
built from the source cleanup revision or later. The image digest pinned by the
native handoff in PR #186 (`d2cdd6141583e857f8509a3a278f4b316be1fd9edf80148dee275d4454323d0a`)
predates this change and must not be assumed to contain it. Before running a
recovery Job, verify that the chosen image contains `/migrate-recovery` and the
exact SQL migration set for the failed release. Publish and pin a reviewed image
with both before using this example. Do not substitute an arbitrary newer
migration image.

Use the same database owner Secret and migration source as the native migration
Job `tdk-prod-tadoku-api/tadoku-api-migrate`:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: tadoku-api-migration-inspect
  namespace: tdk-prod-tadoku-api
spec:
  backoffLimit: 0
  template:
    spec:
      automountServiceAccountToken: false
      restartPolicy: Never
      containers:
        - name: inspect
          image: ghcr.io/tadoku/tadoku/tadoku-api-migrations:prod@sha256:REPLACE_WITH_VERIFIED_DIGEST
          command: ["/migrate-recovery"]
          args:
            - "-source"
            - "file:///migrations"
            - "inspect"
          env:
            - name: POSTGRES_HOST
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-host
            - name: POSTGRES_PORT
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-port
            - name: POSTGRES_DATABASE
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-database
            - name: POSTGRES_SSLMODE
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-sslmode
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-user
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: tadoku-planetscale-admin
                  key: postgres-password
```

Expected dirty output:

```text
version=13 dirty=true
```

Cross-check it directly:

```sql
select version, dirty from schema_migrations;
```

If the two results disagree, stop and investigate the database URL,
`search_path`, and migration table configuration.

## 4. Determine the physical database state

A dirty version means migration `V` started but did not complete cleanly. It
does not prove whether zero, some, or all statements took effect.

Compare the failed `V.up.sql` with:

- tables, columns, constraints, indexes, functions, and types;
- affected row counts and data invariants;
- migration logs and PostgreSQL logs;
- the pre-migration backup.

Classify the database into exactly one state:

| State | Required action |
| --- | --- |
| Migration rolled back completely | Verify the schema matches the previous successful version. |
| Migration completed physically | Verify every intended schema and data effect. |
| Migration partially applied but repairable | Apply reviewed SQL to finish it or restore the previous physical state. |
| Physical state or data integrity is uncertain | Restore the database from backup/PITR. Do not use `force`. |

Do not select a target metadata version until the physical state matches that
version exactly.

## 5. Repair the physical state

The preferred options are:

1. finish the failed migration manually so the database exactly matches `V`;
2. reverse only the known partial effects so it exactly matches the previous
   successful version;
3. restore from the pre-migration backup or PITR point.

Save reviewed repair SQL in the incident record before execution. Run it in an
explicit transaction when PostgreSQL supports transactional execution for all
included statements.

After the repair, repeat all schema and data checks before changing migration
metadata.

## 6. Repair migration metadata

Only after the physical database has been verified, run `force`. The command
requires:

- the dirty version observed by `inspect`;
- the target version matching the verified physical schema;
- a second copy of the target version as explicit confirmation.

For a dirty version `13` whose physical changes were completely reverted to
version `12`:

```text
/migrate-recovery \
  -source file:///migrations \
  -expected-version 13 \
  -target-version 12 \
  -confirm-target-version 12 \
  force
```

Set `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DATABASE`, `POSTGRES_SSLMODE`,
`POSTGRES_USER`, and `POSTGRES_PASSWORD` in the command environment. The
command rejects both `POSTGRES_URL` and the removed `-database` URL flag so
credentials cannot be supplied as a complete URL.

The command refuses to run if:

- the database is not dirty;
- the observed dirty version is not `13`;
- the target version is newer than the observed dirty version;
- the confirmation does not match the target;
- any guard is omitted.

The command reads the metadata again after `force` and fails if the target is
not clean. Run a separate `inspect` Job as an independent check. Expected
output:

```text
version=12 dirty=false
```

Retain the recovery Job and logs as evidence until the incident is closed.

## 7. Ship a forward fix

1. Never edit a migration that succeeded in another environment.
2. Add a new forward migration that is safe from the verified database state.
3. Test first-run, dirty-recovery, forward-fix, and no-op behavior against a
   disposable PostgreSQL database.
4. Publish a reviewed Tadoku API migrations image containing the corrected migrations.
5. Manually sync the Application while automated sync remains disabled.
6. Verify:
   - the migration Job succeeds;
   - `schema_migrations.dirty` is `false`;
   - the Deployment rolls out only after migration success;
   - API and data-integrity smoke checks pass;
   - a subsequent migration run reports `no change`.
7. Restore image updates and automated sync only after the soak period.

## Migration authoring requirement

PostgreSQL migrations must use explicit `begin` and `commit` when every
statement in the migration supports transactional execution. Migrations that
cannot be transactional must document:

- why;
- possible partial states;
- detection queries;
- compensating or completion SQL;
- backup and restore expectations.

Test both failure and recovery paths before production adoption.

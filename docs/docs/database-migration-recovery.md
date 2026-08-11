---
title: Database migration recovery
---

# Database migration recovery

This runbook applies when an API image's Argo CD `PreSync` migration Job fails
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
3. Confirm that the failed PreSync Job did not deploy the new API image.
4. Confirm that the previous API Deployment is available and passes basic
   health checks.
5. Confirm that no migration or recovery Pod is currently running.

Do not start another Argo sync while recovery is in progress.

## 2. Preserve evidence and backup

Capture:

- the failed Job YAML and Pod logs;
- the Application operation result;
- the API image digest;
- the migration filenames embedded in that image;
- the current database backup status and PITR recovery point.

Create an incident-specific backup before modifying the database. Verify that
the backup can be listed or restored in an isolated environment.

## 3. Inspect migration metadata

Run the failed release's exact API image with `/migrate-recovery`. Use the same
database owner Secret and migration source as the PreSync Job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: content-api-migration-inspect
  namespace: tdk-prod-content-api
spec:
  backoffLimit: 0
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: inspect
          image: ghcr.io/tadoku/tadoku/content-api@sha256:REPLACE_ME
          command: ["/migrate-recovery"]
          args:
            - "-source"
            - "file:///migrations"
            - "inspect"
          env:
            - name: POSTGRES_HOST
              value: io-postgres.shared.svc.cluster.local
            - name: POSTGRES_PORT
              value: "5432"
            - name: POSTGRES_DATABASE
              value: tadoku_prod_content
            - name: POSTGRES_SSLMODE
              value: require
            - name: POSTGRES_USER
              valueFrom:
                secretKeyRef:
                  name: tadoku-prod-content-owner-user.io-postgres.credentials.postgresql.acid.zalan.do
                  key: username
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: tadoku-prod-content-owner-user.io-postgres.credentials.postgresql.acid.zalan.do
                  key: password
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
deprecated `-database` URL flag exists only as a bounded operator escape hatch;
do not use it in Kubernetes workload arguments or automation.

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
4. Publish a reviewed API image containing the corrected migrations.
5. Manually sync the Application while automated sync remains disabled.
6. Verify:
   - the PreSync Job succeeds;
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

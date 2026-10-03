# Supported development acceptance (development cluster only)

Scope: webv2, auth, admin, tadoku-api and tadoku-worker. Token-reflector is base-only.
Paper styleguide is deferred; its local pnpm workflow is independent.

Use dev-cli v0.7.0 or newer and only the `homelab-talos-dev` context. Record
the source/CLI revision, exact command, owner, route and full tenant key, Pod
UID/image, browser screenshot and result outside Git. Retain commands and
sanitized output with the PR report. Do not include credentials or cookies.

## Failure modes and gates

1. Bazel discovery must not select unrelated apps: query reverse dependencies
   for an auth-only file, an admin-only file and shared `packages/ui` code.
   Expect auth only, admin only, and all three frontends respectively.
2. Render `kubectl --context homelab-talos-dev kustomize k8s/dev/base`. Validate with
   kubeconform. Auth/admin ingress must use Envoy, retain TLS and hostnames,
   and preserve the more-specific `/kratos` route to shared Kratos.
3. Base: use a fresh browser profile to log in through account.tadoku.dev.lab
   with a synthetic development account, follow its return URL, and load admin.
   Reject attempted requests to production Tadoku hosts.
4. From a fresh branch with a source-only change, run `dev doctor`, then
   `dev up --owner <unique-owner>`. Confirm only the affected app starts.
   Open `dev url --host account.tadoku.dev.lab /login` or the admin equivalent.
   Verify selected backend headers, host-only cookie and an unselected context
   still seeing base. Keep real authentication: no mocked auth or bypass.
5. Change a visible application string again while the same loop runs. Verify
   browser HMR without navigation/state loss and unchanged Pod UID/image.
   Check source deletion/dependency sync where applicable; surface sync errors.
6. Verify failed Go compilation preserves the old process and corrected source
   recovers. Record five warm samples including all outliers: frontend p95 ≤2s,
   backend edit-to-ready/response p95 ≤10s, alongside local/cluster pressure.

Selecting either Tadoku API or tadoku-worker must start both under one owner,
with the same database and full branch tenant. A worker-only source edit
must refresh only its process; an API-only edit must leave the worker Pod UID
and process unchanged. `dev status` and `dev logs tadoku-worker` must work,
but no worker route or URL may be created. The tenant gates below cover their
ownership and teardown.

## Tenant lifecycle gates

Before startup, confirm the shared base has clean migrations, only canonical
tenant `tadoku/prod`, tenant RLS, scoped unique indexes, the current tenant-aware
API/worker, marked Dev Admin/Reader fixtures and the owner Secret. Do not
reseed the shared base just to make a branch check pass.

Snapshot the ordered `pg_database` names and canonical per-table counts before
the run. Run the following through an authorized owner-role connection, with
psql variable `tenant=tadoku/prod`; record the command and output without its
credentials:

```sql
select format(
  'select %L as table_name, count(*) as rows from %I.%I where tenant = %L',
  table_name, table_schema, table_name, :'tenant'
)
from information_schema.columns
where table_schema = 'public' and column_name = 'tenant'
order by table_name
\gexec
```

Use the same query with each full branch tenant key for cleanup evidence. Set
up any authorized canonical job fixture before the baseline, or trace an
existing canonical job, so worker proof does not add canonical rows during
this run. Keep unrelated canonical writes outside the count boundary.

1. **Frontend-only writes.** On a branch with only a webv2 source change, run
   `dev up --owner <owner-a>` without task flags. Record route A and tenant
   `tadoku/<route-a>`. Prove no new database, no database dependency Job and no
   API/worker overlay. Select the main-host link and sign in through the real
   fixture login. Create a uniquely named log through the browser, reload it
   and verify its stored `logs.tenant` equals A's full key. Its API backend may
   be the base; `X-Dev-Selected` must name A. Verify the branch leaderboard and
   that canonical and another tenant's results remain unchanged.
2. **Separate owners.** Start the same Git branch for owner B, using a separate
   browser profile. Record its distinct route and tenant, select B's main-host
   link, and use the same fixture identity. A's log ID must return 404 for B;
   B's seeded data remains readable. Do not substitute another tab in A's
   profile for an independent profile.
3. **Worker ownership.** Start an API branch and confirm its paired worker,
   shared database and full `API_BRANCH` / `WORKER_BRANCH` key. Verify the
   `(tenant, component='tadoku-worker')` override exists before the branch
   worker starts. Trigger a branch write and prove its job completes on the
   branch worker using job ID, persisted tenant and scoped worker logs. Prove a
   canonical job completes on the base worker and that branch processing does
   not invalidate canonical caches. Frontend-only A has no override and its
   jobs complete on the base worker.
4. **Access permits.** With the main hostname selected for the branch, the
   same readable API operation must return 403 anonymously and 200 for Dev
   Reader and Dev Admin. Record `X-Dev-Selected` on each actual API response.
   Reader's access comes from its direct tester tuple; Admin inherits the
   canonical parent's permits. Do not inject identities or bypass login.
5. **In-flight teardown.** Queue jobs through frontend-only A, then immediately
   run `dev down --owner <owner-a>` while work is due or in flight. Prove the
   base API/worker remain Ready and the base worker remains healthy after the
   tenant's jobs cascade away. Record the queue state, teardown result and
   readiness. A failed teardown must retain its marker for diagnosis and retry.
6. **Complete cleanup.** After down, prove all owned overlays and pods are gone,
   the worker override and tenant row are absent, and the counting query
   returns zero for every tenant-owned table. Through scoped port-forwards,
   Keto `GET /relation-tuples?namespace=app&object=tadoku/<route>` must be empty
   and Flipt
   `GET /api/v2/environments/test/namespaces/tadoku_<route>` must return 404.
   Run a scoped Valkey scan; it must find no keys:

   ```sh
   kubectl --context homelab-talos-dev -n tdk-dev-data exec deploy/valkey -- \
     valkey-cli --scan --pattern 'tenant:tadoku/<route>:*'
   ```

   Confirm disposable Jobs and the lifecycle marker are removed last, another
   owner's route remains available, deleted selections fall back to base, and
   canonical providers remain unchanged. Keep tuple lists, key scans and status
   codes in the report; do not replace them with a summary assertion.
7. **Expiry cleanup.** Use an external override configuration with `ttl: 2m`
   and a new owner. Run `dev up --no-watch`, record its expiry, and let that
   owned route expire without a heartbeat. `dev status` must report
   `teardownPending` without running its hooks. Inventory expired routes before
   `dev cleanup`, which can affect other owners, then run cleanup with
   compatible task configuration. Repeat every cleanup check in gate 6 for
   this route; no extra soak is required beyond actual expiry.
8. **Canonical preservation.** Tear down remaining owned fixtures, repeat the
   canonical count query and database-name inventory, and compare them to the
   baseline. Every per-table count and database name must match. Also record
   `select count(*) from pg_database where datname ~ '^tadoku-'`; it must be
   unchanged across the run. Confirm canonical tenant/provider state and the
   base API/worker remain healthy.

All eight gates must pass with inspectable evidence. If provider teardown
fails, retain the marker and tenant row, inspect the recorded task output and
retry. Never force-delete a lifecycle marker, task Lease or database to make
cleanup appear complete.

## Migration branch gates

Use an uncommitted scratch migration on an owned local branch; never merge it
or run it against the shared base. Record the base migration version and
database-name inventory before startup.

1. Run `dev up --owner <owner> --service tadoku-api`. A migration-only change
   needs that service flag because migration files select no deployable.
   Require `profile=isolated-database`, the paired API/worker and the same
   resolved `tadoku-<route>` database and `tadoku/<route>` tenant. The database
   dependency, migration and tenant hooks must complete before startup.
2. Query the isolated database through an authorized owner connection. Require
   owner `tadoku_owner`, marker `dev-cli branch database route=<route>`, clean
   `schema_migrations` at the scratch version, and its registered test tenant.
   Prove non-owner runtime DML/sequence access and denied DDL/migration-table
   writes. Create and reload a browser log; verify its tenant and ID exist in
   this database and not in the base. Shared schema/version and canonical
   per-table counts must remain unchanged.
3. Remove only the owned scratch file while the route remains active. Rerun up;
   require the profile-switch refusal and `run dev down first`, with no new
   lifecycle task or overlay mutations. The marker must retain its original
   profile and resolved database.
4. With the scratch file still absent, run `dev down --owner <owner>`. It must
   use the recorded isolated database, remove its tenant/override and cascading
   rows, clear Keto tuples and Valkey keys, return Flipt namespace 404, and
   remove owned overlays and marker. Reuse tenant gate 6's exact commands with
   the count query connected to the isolated database. The marked database and
   scratch schema remain retained; never drop it as verification cleanup.
5. Run ordinary up for the same owner and require `profile=default` with
   `DATABASE=tadoku`, then normal down. The retained database must remain,
   canonical rows/schema/providers must match baseline, and the base API/worker
   must remain Ready. Include the retained name in any later removal inventory;
   that requires separate exact-name approval.

The offline E2E must also pass its profile/hook declarations, branch-only
migration target, owner/runtime grants and legacy-owner refusal checks. Keep
its report and the live commands, database owner/marker/version, persisted log,
refusal output and cleanup scans outside Git with the PR evidence.

Branch selection is per hostname, not an authentication cookie. For an admin
or auth overlay making API writes, visit both its CLI link and the main-host
link in the same browser profile. Check `X-Dev-Selected` on the API response.
Clearing one host does not clear others.
Run browsers with Lab CA trust; note any TLS verification bypass separately.

These gates cover the supported dev-cli workflow. Repeat
them for changes to routing or synchronization; publish CLI releases only after
their live gates pass. Perform these checks within the authorized development
scope. Named retained-database deletion requires separate explicit approval.

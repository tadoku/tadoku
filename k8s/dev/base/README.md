# Tadoku base — development only

This Kustomize root is exclusively for `homelab-dev` (ct190). It is **not the
production deployment**. Production remains in `antonve/tadoku-argocd` and is not
changed by these manifests. Do not apply this root to another Kubernetes context.

## Ownership and topology

Argo CD owns this always-running base. DevCLI owns only branch overlays, routing
overrides, short-lived tasks and owner/branch database provisioning. Tilt must not
run against this environment. Homelab contains the Application and development
Image Updater infrastructure, not copies of these workload manifests.

Production's `tdk-prod-*` service boundaries become `tdk-dev-*`: the three
frontends, native Tadoku API, remaining immersion/content/profile APIs, Kratos,
Keto, Oathkeeper, Flipt and token-reflector. `tdk-dev-data` contains one
operator-managed Postgres server plus disposable Valkey and Mailhog;
`tdk-dev-routing` attaches application routes to the existing platform Envoy
Gateway. Retired authz/memory/echo services and optional styleguides/admin tools
are not deployed. There is no production data, PlanetScale, Upstash, external
backup or notification configuration here.

Existing CI publishes the nine GHCR runtime/migration images. The root's `images`
entries select `latest`; development Image Updater uses the **digest** strategy
and writes immutable resolutions back to this Kustomization. Do not add another
build/push pipeline. Hook migration images need `force-update` because successful
Jobs are removed from the live resource list. Kustomization-only commits do not
match the existing production image-publication workflow path filters.

The existing Next standalone images bake production URLs into `server.js`.
`frontend-start.cjs` uses their unchanged compiled assets and build configuration,
but supplies development-only runtime endpoints before starting Next. It is a
deployment adapter, not a source-sync server or reusable CLI runtime. Its private
Next startup API must be reverified when the frontend framework version changes.
The public Lab CA is mounted for server-side HTTPS. Branch pods still use the
CLI-managed pnpm/Next dev server and do not use this adapter.

## Automatic migrations

Full Argo syncs execute these waves:

| Wave | Resources and gate |
| --- | --- |
| -30 | Development namespaces |
| -20 | Postgres CR and generated configuration |
| -10 | Tadoku, Kratos and Keto migration Sync hooks; each waits for authenticated database connectivity |
| 0 | Auth providers, cache, Flipt, token-reflector and Gateway routes |
| 10 | Oathkeeper (publishes JWKS before APIs start) |
| 20 | Native and remaining legacy APIs |
| 30 | Frontends |
| 50 | Browser Ingresses |

Migration Jobs have bounded deadlines, `backoffLimit: 0`, and
`BeforeHookCreation,HookSucceeded`. Failed Jobs remain for diagnosis and prevent
the later application wave from starting. Existing healthy application replicas
are not deleted on failure. No-op reruns succeed. This does not authorize a
database reset, schema downgrade or rollback. Fix forward and perform another
full sync. Never use selective resource sync to release an application: it skips
Argo hooks. Do not configure `ApplyOutOfSyncOnly=true`.

Use the repository's migration-first release contract: schema changes land and
deploy independently before dependent runtime code. Digest tracking is not a
substitute for backward-compatible schema/application releases.

Branch `make dev-up` invokes `dev up --task migrate --task seed`; the dependency
creates its database first and task failure prevents overlay startup. Plain
`dev up` does not implicitly run tasks. `dev task migrate` and `dev task seed`
remain explicit, non-destructive reruns. Base seeding is intentionally separate
from migrations: `make dev-seed` uses the existing marked synthetic identities.

## Credentials and initial bootstrap

There are no plaintext Secret manifests or private keys in this root. The
Postgres operator generates the `immersion`, `kratos` and `keto` credentials in
`tdk-dev-data`. `scripts/dev/bootstrap-gitops-secrets.sh` copies only required
credentials into consumer namespaces, generates development-only signing/session
material once, and preserves existing keys on rerun. Run it only with explicit
credential/bootstrap approval. It fails closed on another API server or unowned
namespaces and never reads production or old Tilt credentials.

The normal initial sequence is: approve the cutover, land this path on main,
configure development repository access, then start a **full** Argo sync. Once
namespaces and operator-generated Secrets exist, run:

```sh
bash scripts/dev/bootstrap-gitops-secrets.sh
# Wait for migration Jobs and base workloads to become Healthy, then:
make dev-seed
```

The initial sync can wait at the auth-provider wave until bootstrap supplies
runtime Secrets. Rerun bootstrap when operator credentials rotate, then perform
an explicitly approved restart of consumers. Kubernetes does not refresh env
variables in existing processes. Base migrations require no manual task on
subsequent full syncs. Keep generated keys only in development Kubernetes Secrets
or an approved encrypted backup, never in Git or this worktree.

One shared Postgres pod serves base databases and `immersion-<owner/branch-route>`
branch databases. The admin Secret stays in `tdk-dev-data` and is used only by the
short-lived branch database creation Job. Application pods and tasks use the
`immersion` role. This is cooperative isolation, not hostile multi-tenancy.
Postgres CR and data namespace have `Prune=false,Delete=false`; ordinary Argo
pruning or Application removal must not destroy data. Branch databases are also
retained on `dev down`; deletion needs exact ownership checks and permission.

## Cutover is not adoption

The owner-approved fresh cutover was activated on 2026-09-21 through Homelab
#430. The old Tadoku namespaces, canonical routes, `default/tadoku-dev-db`, its
PVC and the old Valkey PVC were deleted as explicitly approved disposable data.
No backup was retained. Unrelated development workloads and production were not
deleted or adopted. New runtime credentials were generated in development.

This is a fresh design derived from main's Tilt configuration and production's
topology, not exported live objects. **Do not activate it alongside the old
canonical Ingresses and HTTPRoutes.** Both claim the same hosts and the older
Gateway route can continue winning traffic selection. Bootstrap also creates a
fresh database and auth identities: existing sessions will not transfer.

Before activation, inventory and approve the exact old route/Ingress replacement
and the disposition of the old `default/tadoku-dev-db` and its branch databases.
Do not run `tilt down`, delete old namespaces or delete database PVCs as an
implicit cleanup step. A temporary second base during a specifically reviewed
cutover is not the steady-state database model. The intended final state has one
Postgres pod. Preserve old data until its deletion is separately authorized.

These gates also apply to any future fresh cutover; the completed one-time
deletion approval is not blanket permission to erase later data.

## Verification

```sh
kubectl --context homelab-dev kustomize k8s/dev/base |
  kubeconform -strict -summary -ignore-missing-schemas
node k8s/dev/base/e2e.cjs /tmp/tadoku-base-evidence
```

The E2E command uses real published migration images and isolated, resource-bounded
Docker containers—not a nested Kubernetes cluster. It records source revision,
worktree status, rendered-manifest hash, exact image digests, commands and results
in `report.json`, with individual logs. It proves fresh migrations, no-op reruns,
an intentional connection failure and recovery, plus HTTP readiness, development
runtime URLs and compiled assets in all three published frontend images with
external networking disabled. It then cleans up only its labeled fixtures.
Install the frontend's locked pnpm dependencies first (the harness reuses its
YAML parser). Synthetic credentials in fixture logs are not live credentials.

This offline gate **does not prove** Argo sequencing, a failed schema migration
blocking a live rollout, Image Updater commits, ingress/authentication or branch
isolation. Those remain mandatory live acceptance gates after activation. Verify
the actual browser login and leaderboard, two owners, independent frontend/API
fallback, HMR, Go binary replacement, migration failure gating and scoped cleanup.
Retain browser traces/screenshots outside source control and identify the exact
tested revision and any substituted routing boundary.

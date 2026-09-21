# Tadoku dev-cli pilot

This is an opt-in pilot for **webv2 (Next.js/pnpm)** and **native tadoku-api**.
Tilt still owns the shared development stack. This does not retire Tilt, move
base workloads to Argo CD, or change production deployment.

## Boundaries

- Bazel owns deployable discovery, runtime images and application binaries.
  `dev_deployable` dependencies include webv2's shared `ui` package.
- dev-cli owns branch overlays, image publication/digest injection, live sync,
  its supervised binary runtime, routes, task Jobs and dependency lifecycle.
  There is no Tadoku-specific development router or supervisor implementation.
- An existing Envoy Gateway handles cookie/header routing. The extra pilot
  Oathkeeper instance uses the existing development authentication providers
  and signing Secret by reference. It does not create credentials.
- Databases and tasks are **explicit**. Shared databases are the default.
  An isolated native API can still proxy to unchanged legacy APIs using shared
  data. No automatic service groups or consistency restrictions are imposed.
- This is isolation for cooperative developers, not an authentication or tenant
  security boundary. Shared Kratos, Keto, Valkey and legacy APIs remain shared.

Browser → ingress-nginx → Envoy → webv2 or pilot Oathkeeper → Envoy → native API.
Oathkeeper preserves normalized branch context on its internal hop. The native
API retains its ordinary legacy proxies. Cookie-derived selection takes
precedence over browser-supplied `x-dev-branch`.

## Prerequisites and local configuration

Use CLI commit `9ab8ef8517d7a1745f431ad749a56b890e1772e1`
or a release containing it; older releases do not have dependency/task support:

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@9ab8ef8517d7a1745f431ad749a56b890e1772e1
cp .dev/config.example.json .dev/config.json
mkdir -p .dev/local
cp .dev/pilot.example.json .dev/local/pilot.json
```

Keep cluster-specific values in these ignored files. Configure the development
context, authorized namespaces, registry, unique pilot hostname and existing
Gateway references. In the pilot manifest replace example hostnames, Gateway
names/namespaces and Service references, and insert the **public** development
CA certificate. Never copy private keys, Secret contents or kubeconfig into Git.

Build and publish the clean webv2 source image, then replace the base
Deployment's invalid image with the immutable registry digest. Do this after
the cluster-specific edits above so rendering does not overwrite them:

```sh
pilot_registry=$(jq -er '.registry | select(type == "string" and length > 0)' .dev/config.json)
pilot_repository="${pilot_registry}/webv2"
pilot_tag="pilot-base-$(git rev-parse --short=12 HEAD)"
# This pilot registry is anonymous; do not reuse a read-only admin Docker config.
pilot_docker_config=$(mktemp -d)
DOCKER_CONFIG="$pilot_docker_config" bazel --host_jvm_args=-Xmx1024m run --jobs=1 //frontend:webv2_dev_push -- \
  --repository "$pilot_repository" --tag "$pilot_tag"
pilot_digest=$(DOCKER_CONFIG="$pilot_docker_config" docker buildx imagetools inspect \
  "${pilot_repository}:${pilot_tag}" --format '{{json .Manifest.Digest}}' | jq -er .)
pilot_image="${pilot_repository}@${pilot_digest}"
jq --arg image "$pilot_image" '
  (.items[]
    | select(.kind == "Deployment" and .metadata.name == "tadoku-cli-web-base")
    | .spec.template.spec.containers[]
    | select(.name == "webv2")
    | .image) = $image
' .dev/local/pilot.json > .dev/local/pilot.rendered.json
mv .dev/local/pilot.rendered.json .dev/local/pilot.json
jq -e --arg image "$pilot_image" '
  any(.items[];
    .kind == "Deployment"
    and .metadata.name == "tadoku-cli-web-base"
    and any(.spec.template.spec.containers[]; .name == "webv2" and .image == $image))
' .dev/local/pilot.json >/dev/null
```

The shared Tilt stack must already supply the native API base Service, Ory
providers, Valkey, Postgres operator and referenced Secrets. The pilot manifest
owns its base webv2 Deployment and Service; their selector is isolated from the
shared Tilt frontend. The native API's shared schema must already match its
binary; this pilot never migrates the shared database. The gateway must support
Envoy `Backend` and Gateway API resources, and permit routes from the separately
labeled pilot namespace. Review every manifest target before applying; do not
relabel shared namespaces or modify the Gateway to bypass a policy rejection.

```sh
kubectl --context "$DEV_CONTEXT" apply --dry-run=server -f .dev/local/pilot.json
kubectl --context "$DEV_CONTEXT" apply -f .dev/local/pilot.json
dev doctor
```

`pilot.example.json` is an operator-reviewed bootstrap example, not an automatic
CLI apply hook or another Tilt entrypoint. All its resources have the dedicated
`app.kubernetes.io/managed-by=tadoku-dev-cli-pilot` label. The shared base remains
under Tilt; adopting static pilot configuration into GitOps is a later phase.

## Shared-data development

Make a service edit before startup. Unknown files conservatively select both
deployables; known source changes follow Bazel reverse dependencies. Initial
integration changes affect both. Use an explicit comparison ref when testing a
single-service edit on top of this integration commit.

```sh
dev up --owner alice
# Keep that terminal running. In another terminal, same checkout:
dev status --owner alice
dev url --owner alice '/desired/path'
dev logs --owner alice tadoku-api
```

Open the printed link; it sets a host-only branch cookie without application UI
changes. For the normal Navbar `Log in` flow, an operator must temporarily add
the exact pilot origin to shared development Kratos's return-URL and CORS
allowlists, preserving existing origins and avoiding wildcards. This is a
live-only acceptance prerequisite, not an auth configuration change for this
PR. Each browser profile has its own branch selection; tabs within one profile
share it. Signing in at the account hostname first remains an optional manual
diagnostic, but it does not satisfy the full acceptance gate.

Source edits sync into the existing pnpm Next dev server. Go edits rebuild only
the selected binary and restart it in the same pod after upload. A compilation
failure leaves the working binary serving and is visible in `dev status`.
Add another deployable explicitly with `--service` on a restarted loop; the CLI
does not continuously expand its initial affected-service selection.

The pilot-owned base frontend uses the pilot hostname for SSR and a relative
`/api/internal` browser endpoint. An API-only overlay therefore falls back to a
frontend that remains on the pilot origin and sends its branch cookie through
the pilot Oathkeeper route. The shared Tilt frontend and its canonical runtime
configuration are unchanged. Both pilot Next.js processes inherit a 384 MiB
V8 heap ceiling from `tadoku-cli-web` and have a 1 GiB container limit so a base
and one overlay remain bounded on the development node.

## Disposable native database

In a separate branch/worktree with a distinct owner, append
`--define=dev_cli_database=isolated` to `bazelArgs` in the ignored configuration.
This selects the isolated API template; it does **not** provision anything.
Explicitly create the database and run existing migrations and the tiny fixture:

```sh
dev up --owner bob --task migrate --task seed
# Explicit, serialized reruns; flags must precede the task name:
dev task --owner bob migrate
dev task --owner bob seed
```

Both tasks depend on the `postgres` declaration. The operator creates a
`tadoku-cli-${DEV_ROUTE}` cluster in `tdk-tadoku-api`, with its own credentials
and local-path storage. The CLI builds/pushes task images and injects immutable
digests. The seed accepts only a `tadoku-cli-*` database host and upserts two
fixed fixture IDs; it does not modify Kratos/Keto or any shared data. No schema
migration is introduced by this pilot. The fixture is readable at
`/api/internal/content/pages/main/dev-cli-pilot` on the selected environment.

Migrations and seed share a target lock. Never delete a Lease to force a retry.
A failed task blocks that startup and never resets a database. See dev-cli's
task recovery documentation for proving an abandoned holder has stopped.

## Acceptance

Use two owners/worktrees: A with webv2 + API/shared DB, B with API-only/isolated
DB. For a temporary observable marker, add
`w.Header().Set("X-Tadoku-Pilot", "<distinct-marker>")` in `Router.ServeHTTP`
in each checkout. Give A's `frontend/packages/ui/styles/globals.css` a temporary
`body { background-color: rgb(219, 234, 254) !important; }` rule. Do not commit
these application edits. Make another edit while each original loop runs.

The browser check uses Playwright with its system dependencies (including fonts)
and two **existing** development identities
(administrator and non-administrator). Supply credentials through environment
variables; it creates sessions, never accounts or role assignments. First verify
the pilot/account endpoints with normal TLS-validating curl. Browser contexts
are ephemeral and ignore certificate errors only because their profiles lack
the host's private CA trust store. Do not record storage state or cookies.

```sh
# Supply PILOT_URL, AUTH_URL, ROUTE_A, ROUTE_B, MARKER_A, MARKER_B,
# ADMIN_EMAIL, ADMIN_PASSWORD, READER_EMAIL and READER_PASSWORD in your shell.
# PLAYWRIGHT_MODULE optionally points to an existing Playwright module.
node .dev/acceptance.mjs
# Narrow routing diagnosis when existing login is unavailable (NOT a full pass):
node .dev/acceptance.mjs --routing-only
# Check that the normal UI still renders data after all overlays are removed:
node .dev/acceptance.mjs --base-only
```

The full script starts on each selected pilot route, follows the real Navbar
`Log in` link to the account form, requires its `return_to` to bring the browser
back to that pilot route, and reads the resulting Kratos session from the pilot
page to prove the exact CORS origin. It then verifies authenticated SSR, native
authorization, legacy proxy traffic, two selections, spoofed routing headers,
isolated seed visibility, actual leaderboard rendering on both overlay and base
frontend, and branch switching/clearing. A blank-page input control distinguishes
a broken browser installation from an application login failure. The live
fixture currently requires HTTPS development Lab hosts. It does not weaken
provider authentication or retry guessed credentials. Independently test a
deliberate Go compilation failure, old-response availability, recovery, pod
UID/image stability, and browser HMR without navigation. Measure from file
modification to visible browser update or successful new API response; report
sample counts and ranges.

Repository checks before review:

```sh
bazel mod deps --lockfile_mode=error
bazel run //:gazelle -- -mode=diff
bazel build //services/...
# Point test variables at disposable LOCAL PostgreSQL/Valkey, never shared data.
bazel test //services/...
bazel build //frontend:webv2_dev_image //frontend:webv2_dev //services/tadoku-api:dev //.dev:seed_image
cd frontend
pnpm install --frozen-lockfile
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 lint
pnpm build
```

## Cleanup and delivery

```sh
dev down --owner alice
# Run in Bob's original checkout:
dev down --owner bob
```

Verify the other owner's routes and the base continue serving when one owner
stops, and a selected deleted overlay falls back to base. `down` stops local
loops and removes that owner's overlays/tasks, but the database declaration uses
`retain`: Postgres and its storage intentionally survive down and TTL cleanup.
Before deleting disposable data, inspect the exact CR/PVC names and CLI ownership
metadata, confirm no consumers remain, and explicitly authorize their deletion.
Never delete the shared `tadoku-dev-db` or its credentials. TTL cleanup runs only
when invoked by a CLI process; there is no permanent overlay reaper/controller.

Keep the static pilot until its operator chooses to remove it. If removing it,
review the rendered file and ownership labels first; do not delete the namespace
while overlays or retained dependency references still exist.

This phase should land as one Tadoku PR. A dev-cli fix, if needed, belongs in a
separate CLI PR. **Tadoku's existing main workflow publishes production images
when these Bazel paths change.** A passing pilot does not authorize that normal
publication/rollout; resolve that boundary before merging. No production image
targets or workflows are changed here.

# Admin migration from Next.js to TanStack Start

## Status and prerequisite

Planning only. **Do not start implementation until the design-system rework is
complete and the resulting `ui` package contract has been adopted by the admin
application.**

The prerequisite is satisfied when:

- the reworked design system is merged and available on `main`;
- the admin application builds against it without migration-specific shims;
- buttons, forms, layouts, toasts, and other shared components used by admin no
  longer depend on Next.js runtime APIs; and
- the design-system owners confirm that its public API is stable enough for the
  duration of the router migration.

This plan deliberately does not add an end-to-end test environment. The admin
application has one operator, its API environment is expensive to reproduce,
and privileged API operations enforce authorization independently of the
frontend. Validation consists of build and type checks, local production-server
smoke checks, and a short manual production checklist with a digest-based
rollback available.

## Overview

Replace Next.js Pages Router in `frontend/apps/admin` with TanStack Start and
its file-based router while preserving the existing URLs, user interface, Ory
session behavior, API clients, and Kubernetes deployment contract.

This is a framework and routing migration, not an admin redesign. Existing
page components and React Query hooks should be reused wherever practical.
TanStack route loaders should initially handle only session bootstrap, route
guards, redirects, and runtime configuration; moving data fetching to loaders
or upgrading React Query is separate follow-up work.

### Scope

- TanStack Start with file-based routes and a Node container runtime.
- Existing dashboard, users, languages, announcements, pages, and posts routes.
- Request-scoped Ory session bootstrap and admin-route protection.
- Runtime configuration supplied by the existing Kubernetes environment.
- Kubernetes startup, readiness, and liveness endpoints and probes.
- Production cutover with manual validation and digest-based rollback.
- Final removal of the legacy `NEXT_PUBLIC_*` environment-variable names.

### Non-goals

- Redesigning admin screens or changing the design-system API.
- Changing backend APIs, Ory, Keto, or the authorization model.
- Migrating React Query v3 or rewriting existing API hooks as route loaders.
- Introducing React Server Components or moving mutations to server functions.
- Building a new end-to-end test environment.
- Migrating the other Next.js applications.

## Benefits

- Typed route parameters, search parameters, links, and navigation.
- Nested and pathless layouts replace the custom `getLayout` convention.
- Route-level loading, error, and not-found boundaries.
- Vite-based development and a smaller framework surface for this client-heavy
  internal application.
- A contained production trial before considering any other application.
- Explicit request-scoped session and query state instead of module-scoped SSR
  state.

## Risks and mitigations

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Ory cookies or return-to redirects behave differently | Login, logout, or deep links fail | Resolve the session from the incoming request on the server; test signed-out, signed-in, expired-session, and direct-link flows manually before cutover |
| Browser-only editor modules fail during Vite SSR | Page and post editors fail to render | Exercise CodeMirror routes in the first spike and isolate browser-only code behind client-only boundaries if required |
| Runtime values are accidentally compiled into the client build | Local configuration fails or an internal service address is exposed | Read environment variables on the Start server and serialize an allow-listed public configuration object from the root loader |
| Start/Nitro output differs from the shared Next.js image layout | Production container does not start | Add an admin-specific image path or parameterize the shared Dockerfile without changing other Next.js images |
| A probe is enabled before its endpoint exists | Kubernetes continually restarts the old image | Ship and verify health endpoints first; add probes in `tadoku-argocd` only after that image is deployed |
| A route or mutation regresses without E2E coverage | The sole admin operator is temporarily blocked or sees incorrect behavior | Keep existing API hooks, run a focused manual checklist, and retain the previous image digest for immediate Git-based rollback |
| TanStack Start release-candidate or Nitro changes cause churn | Dependency upgrades break builds or runtime behavior | Pin exact framework versions for the migration and upgrade deliberately after cutover |

## Measurable success criteria

- All existing admin URLs resolve with the same path parameters and redirect
  behavior, including direct navigation and browser refresh.
- Signed-out visitors are redirected to Ory with the original admin URL as the
  return target; signed-in non-admin users do not see admin content.
- The dashboard and one create/edit flow for each content family can be used in
  production without changing backend APIs.
- Browser bundles contain no Kratos internal endpoint or other server-only
  configuration.
- The production container listens on port 3000 and all three health endpoints
  return HTTP 200 in their healthy state.
- Kubernetes uses `startupProbe`, `readinessProbe`, and `livenessProbe`, and a
  rolling deployment keeps the previous pod available until the new pod is
  ready.
- `pnpm --filter admin exec tsc --noEmit`, the admin build, and the production
  container smoke check pass.
- Production manual validation passes, or the deployment is restored by
  reverting to the previously recorded image digest.
- No `NEXT_PUBLIC_*` variables remain in the admin source, build workflow,
  development manifest, or production ArgoCD manifest after the final phase.

## Phase 0: prerequisite and migration baseline

- [ ] Confirm the design-system rework is complete using the prerequisite
  criteria at the top of this document.
- [ ] Rebase the migration branch on the `main` commit containing that work.
- [ ] Inventory the admin route paths, redirect-only pages, `next/*` imports,
  runtime configuration consumers, Ory session behavior, and Docker assumptions.
- [ ] Record the currently deployed admin image digest from `tadoku-argocd` as
  the rollback target.
- [ ] Write a short manual validation sheet covering login, logout, dashboard,
  direct dynamic-route refresh, and a harmless create/edit operation.

**Gate:** the design-system prerequisite is complete, the admin builds on
`main`, the current route/configuration inventory is recorded, and a known-good
production image digest is available. No migration work starts before this
gate passes.

## Phase 1: prove the Start runtime

- [ ] Replace the admin package's Next.js scripts and configuration with a
  minimal TanStack Start/Vite setup while retaining React 18 unless a concrete
  dependency requires otherwise.
- [ ] Upgrade TypeScript for the admin package to a TanStack-supported version
  and update its module-resolution settings.
- [ ] Add the generated file-route tree to development and production builds.
- [ ] Create the root document, global providers, stylesheet imports, default
  metadata, not-found boundary, and error boundary.
- [ ] Migrate the dashboard and one dynamic editor route as a vertical slice.
- [ ] Verify CodeMirror, DOMPurify, local storage, and other browser-only
  dependencies under development SSR and a production build.
- [ ] Keep existing React Query hooks and API request behavior unchanged.

**Gate:** the dashboard and representative editor work through the Start
development server and production build, including direct navigation and
refresh, with no design-system fork or backend change.

## Phase 2: session, authorization UX, and runtime configuration

The auth/configuration work and the remaining route-file preparation may be
performed in parallel once Phase 1 has established the project structure.

- [ ] **[Auth]** Move the Ory server SDK into a server-only module and resolve
  the session from the incoming cookie for each request.
- [ ] **[Auth]** Add a pathless protected admin layout that redirects signed-out
  users and denies non-admin users before rendering child routes.
- [ ] **[Auth]** Preserve return-to URLs, logout behavior, expired-session
  handling, and the existing backend authorization boundary.
- [ ] **[State]** Construct session, Jotai, and Query Client state per request;
  do not carry forward module-scoped initialized-session state.
- [ ] **[Configuration]** Read the existing Kubernetes `NEXT_PUBLIC_*`
  variables on the Start server at request time.
- [ ] **[Configuration]** Validate required values at startup and expose only
  the browser-safe subset through the root loader.
- [ ] **[Configuration]** Keep the Kratos internal endpoint server-only and
  confirm that it is absent from generated browser assets.
- [ ] **[Unauthorized flow]** Replace `/api/unauthorized` with a Start server
  route or equivalent server-generated redirect.

**Gate:** login, logout, return-to, non-admin denial, expired-session handling,
and runtime configuration work against the real development environment; the
client build contains no internal Kratos address.

## Phase 3: complete file-route migration

After Phase 2 establishes the protected layout and routing conventions, the
route families below are independent and may be assigned in parallel.

- [ ] **[Dashboard/users]** Migrate `/`, `/users`, and their page metadata.
- [ ] **[Languages]** Migrate `/languages` without changing its API hooks or
  form behavior.
- [ ] **[Pages]** Migrate list, view, create, and edit routes under
  `/pages/$namespace`, including the default-namespace redirect.
- [ ] **[Posts]** Migrate list, view, create, and edit routes under
  `/posts/$namespace`, including the default-namespace redirect.
- [ ] **[Announcements]** Migrate list, create, and edit routes under
  `/announcements/$namespace`, including the default-namespace redirect.
- [ ] Replace `next/link`, `next/router`, and `next/head` with typed TanStack
  navigation, route params, search params, and route metadata.
- [ ] Remove the Pages Router tree, custom `getLayout` convention,
  `next.config.js`, and all remaining admin imports from `next/*`.
- [ ] Confirm the generated route tree is current during typecheck and build.

**Gate:** every inventoried route and redirect has a TanStack equivalent, no
admin source imports Next.js, and the complete admin application passes its
typecheck and production build.

## Phase 4: production container and Kubernetes health checks

Application health endpoints must land before the ArgoCD probe configuration.
The endpoint implementation and Docker work can proceed in parallel, but both
must pass before the ArgoCD change is merged.

- [ ] **[Application]** Add `GET /startupz`; return HTTP 200 after runtime
  configuration has been validated and the Start server is initialized.
- [ ] **[Application]** Add `GET /readyz`; return HTTP 200 when the process can
  serve admin requests. Do not make readiness depend on optional upstream APIs
  in a way that causes cascading pod removal.
- [ ] **[Application]** Add `GET /livez`; return HTTP 200 while the Node process
  and request loop are alive. Do not perform expensive dependency checks.
- [ ] **[Application]** Keep the health responses unauthenticated, minimal, and
  free of configuration or dependency details.
- [ ] **[Container]** Produce a non-root Node runtime image containing the
  Start/Nitro server output and static assets, listening on port 3000.
- [ ] **[Container]** Avoid breaking the Docker builds for `webv2`, `auth`, and
  `styleguide`; use an admin-specific target or explicit framework parameter if
  the shared Dockerfile remains shared.
- [ ] Build and run the production image locally, then verify `/startupz`,
  `/readyz`, `/livez`, `/`, and one static asset.
- [ ] Deploy the endpoint-capable image with the existing ArgoCD manifest and
  verify the endpoints inside the production pod before adding probes.
- [ ] **[ArgoCD, separate repository]** Add an HTTP `startupProbe` for
  `/startupz`, an HTTP `readinessProbe` for `/readyz`, and an HTTP
  `livenessProbe` for `/livez` on port 3000.
- [ ] **[ArgoCD]** Use a generous startup failure window, a responsive readiness
  interval, and conservative liveness thresholds; validate the rendered
  manifest before merging.
- [ ] **[ArgoCD]** Confirm a rolling update retains the old replica until the
  new replica passes readiness.

**Gate:** the production pod passes all three probes, ArgoCD reports the
application healthy, the service remains available during a rolling update,
and reverting the image digest remains a viable rollback.

## Phase 5: cutover and manual acceptance

- [ ] Record the outgoing and incoming image digests before deployment.
- [ ] Deploy during a window when the sole admin operator can immediately
  validate or revert the change.
- [ ] Verify signed-out login redirection and preservation of the return URL.
- [ ] Verify dashboard rendering, navigation, a direct dynamic URL, and browser
  refresh.
- [ ] Verify one harmless create/edit flow and confirm the intended namespace
  and record ID reach the API.
- [ ] Verify session expiry or logout returns to the authentication UI.
- [ ] Inspect pod restarts and startup/readiness/liveness probe status.
- [ ] If validation fails, revert the ArgoCD image-digest update and document
  the failure before attempting another rollout.

**Gate:** the sole operator completes the manual checklist successfully and
the deployment remains healthy. Otherwise the previous digest is restored.

## Phase 6: rename legacy environment variables

Perform this only after the Start migration is stable. Use a two-step
compatibility rollout so the application and ArgoCD changes can be deployed
independently.

| Current name | Final name |
| --- | --- |
| `NEXT_PUBLIC_KRATOS_INTERNAL_ENDPOINT` | `TADOKU_KRATOS_INTERNAL_ENDPOINT` |
| `NEXT_PUBLIC_KRATOS_ENDPOINT` | `TADOKU_PUBLIC_KRATOS_ENDPOINT` |
| `NEXT_PUBLIC_AUTH_UI_URL` | `TADOKU_PUBLIC_AUTH_UI_URL` |
| `NEXT_PUBLIC_ADMIN_URL` | `TADOKU_PUBLIC_ADMIN_URL` |
| `NEXT_PUBLIC_HOME_URL` | `TADOKU_PUBLIC_HOME_URL` |
| `NEXT_PUBLIC_API_ENDPOINT` | `TADOKU_PUBLIC_API_ENDPOINT` |
| `NEXT_PUBLIC_COOKIE_DOMAIN` | `TADOKU_PUBLIC_COOKIE_DOMAIN` |
| `NEXT_PUBLIC_COOKIE_SECURE` | `TADOKU_PUBLIC_COOKIE_SECURE` |

- [ ] Add support for the final names while retaining the old names as
  temporary fallbacks; prefer the final names when both are present.
- [ ] Update the local/Tilt admin deployment template to supply the final names.
- [ ] Update `tadoku-argocd` to supply the final names and deploy it.
- [ ] Confirm runtime configuration, login, and health checks with only the
  final names present in the pod.
- [ ] Remove the fallback reads and all remaining `NEXT_PUBLIC_*` references
  from the admin application and its deployment files.
- [ ] Search both repositories for stale admin references and update relevant
  operational documentation.

**Gate:** production uses only the final names, both repositories contain no
admin `NEXT_PUBLIC_*` references, and the manual login/dashboard check still
passes.

## Follow-up improvements enabled by the migration

- Evaluate upgrading React Query separately and integrating selected queries
  with route loaders and cache hydration.
- Add route-level pending and error experiences where they materially improve
  admin workflows.
- Reuse the health-endpoint and Start-container pattern if another application
  is selected for migration.
- Add a readiness probe to other frontend deployments after confirming the
  operational behavior on admin.
- Revisit exact TanStack Start and Nitro versions after a stable v1 release.
- Remove temporary compatibility documentation after the environment-variable
  rename has been stable for at least one normal release cycle.

---
title: Authv2
description: Paper account screens, native and enhanced JSON submissions, recovery, Settings and account deletion, session reads, manual browser smoke and deployment.
---

# Authv2

Read this when you run or extend `frontend/apps/authv2`.

Authv2 is the account application on the [Webv3 stack](webv3.md):
TanStack Start, Router, Query and Form, Zod, Vite, Nitro and Paper. `/` is
Settings, the identity API's recovery redirect target; `/settings` opens the
same screen for existing caller links. `/login`, `/register` and `/recovery`
log in, create an account and request a password-reset email.
`/delete-account` requests the account's deletion and `/deleted` confirms it.

## Forms and recovery

Each form uses `useAccountForm` from `frontend/apps/authv2/src/lib/account-form.ts`
with Paper's `usePaperForm` controls. The server renders enabled native forms that
post to `/actions/{operation}`, so every form works with JavaScript disabled or
before hydration. On hydration the form adopts values already typed into those
controls, and again from the live controls at submission.

A hydrated form sends JSON from the browser to the identity API on the tenant's
main host, for example `https://tadoku.app/api/internal/identity/login`, with
credentials and the API's CORS rules. Login throttling therefore sees each
visitor's own address. Branch account hosts do not route `/api/internal`. Client
validation mirrors the API's length rules; the API's similarity rule and other
rejections map onto the field they concern (`email_taken`, `invalid_email`,
`invalid_display_name`, `invalid_password`, `confirmation_mismatch`) or onto the
form (`wrong_password`, `throttled`, `stale_session`, `admin_account`,
`banned_account`, a 401 without a code, `invalid_origin` and service failures).
A wrong password clears and focuses the password field. On registration,
`throttled` means the tenant's per-network registration budget is spent, and
the message says to wait 15 minutes; on login it names the login limit. A
registration answer with an undocumented status, such as the API's `504` or
`499`, or a malformed body counts as a service failure.

Both paths send the form through `sendIdentity()` in
`frontend/apps/authv2/src/lib/identity.ts`, which calls the generated identity
operation with only that form's fields and reads error codes with the generated
error validator.

The `/actions/{operation}` server route completes native submissions only. It
accepts a bounded, single-valued form from this host's own pages (the `Origin`
must match the request host, itself an allowed Tadoku app), calls the fixed
identity operation with JSON, forwards Cookie and the trusted `X-Dev-Branch`,
returns the API's Set-Cookie headers and redirects. Failures redirect with a
public error code shown as a form message; passwords and account values never
enter URLs. Native logins reach the API from the authv2 server, so the API's
per-address throttle counts them under that server's address.

Settings shows two groups of read-only rows: Profile holds the display name,
and Sign-in and security holds the email address and password. Below the
groups, one plain, muted **Delete account** link opens `/delete-account` with
the trusted `return_to`; Settings has no other delete control. Each row shows
its stored value and a **Change** link. The `edit` query parameter
(`display-name`, `email` or `password`) opens one editor in place of its row's
value. `settingsView()` in `frontend/apps/authv2/src/lib/search.ts` reads it
with the other Settings parameters: a failed native submission's `form` and
`error` reopen the editor that sent it ahead of any `edit`, and `saved` shows a
row's confirmation only while no editor is open. **Change** and the editor's
**Cancel** are links that keep the path and the trusted `return_to`, so they
work without JavaScript and only one editor is open at a time. The open editor
focuses its input and has the page's only filled button.

Each form sends only its own field. A hydrated save navigates to the same
`?saved=<operation>` URL that a native save redirects to. That navigation reads
the session again, collapses the row and shows its confirmation. Email and
password changes need a login within the last hour; their stale-session
message links to a login that returns to the same Settings URL with that editor
open. Logout posts to the identity API through the native route. A guest
opening Settings is sent to login with Settings as the continuation.

The account's email is shown exactly as the API sends it, which for an imported
address can include control characters. An email input drops line breaks and
surrounding ASCII whitespace, so its value can differ from the stored address.
The email form therefore also carries the stored address in a hidden
`current_email` field. When the submitted address equals the stored one as an
input shows it, `sendIdentity()` sends nothing and reports `unchanged_email` on
the email field, on both the hydrated and native paths. An edited address is
sent as typed.

An incoming `return_to` is external input. `continuation()` in
`frontend/apps/authv2/src/lib/identity.ts` accepts only HTTPS URLs on the
identity API's host allowlist whose longest Settings URL,
`/?edit=password&return_to=…`, fits 2,048 bytes. That budget is 14 bytes more
than a bare `/?return_to=…`, so authv2 never offers a destination the recovery
redirect would refuse. Anything else is ignored and never rendered. Login and
registration continue to it. Settings offers it as **Back to Tadoku** and
follows it after a password change; display name and email changes stay on
Settings. Recovery requests carry it, or the Tadoku home page, into the emailed
link. A recovery redirect that carries `edit=password` lands with the Password
editor open, and with a continuation its button reads **Save password and
continue**. Invalid, expired or reused recovery
links still return the identity API's `400` JSON response; request another at
`/recovery`. Recovery requests show the same confirmation for known and unknown
addresses. Development mail remains held.

## Account deletion

`/delete-account` says what a deletion removes and keeps and that it can't be
undone, and links to the log export on the visitor's Webv2 profile. Its field,
**Type your email address to confirm**, has autocomplete and spellcheck off.
**Delete my account** is the destructive button; **Cancel** returns to Settings
with the trusted `return_to`.

The route's `beforeLoad` runs during server rendering as well as in the
browser, so its redirects work without JavaScript. A guest goes to login and
comes back. A session whose `recent_login` is false goes to `/login?reauth=1`,
continuing to the page and its `return_to`; a `return_to` too long to nest is
dropped, not the return to the page. `/login` with `reauth` and a session shows
"Privileged actions require a fresh login. Log in again to continue." in place
of the signed-in notice and prefills the account's email, also after a failed
native login. A session without `recent_login` reaches the page, and the API's
`stale_session` answer applies.

Before hydration and without JavaScript, the button posts to
`/actions/delete-account`. Once hydrated, it stays disabled until the typed
value matches the account's address, ignoring letter case and surrounding
whitespace; the identity API makes the authoritative comparison. Both paths
send only `email` to `POST /identity/settings/account-deletion` and land on
`/deleted` after its `202`; the native path forwards the API's cookie-clearing
Set-Cookie. A `stale_session`, meaning the hour ran out while the page was
open, or a 401 without a code goes to the same fresh-login URL.
`confirmation_mismatch`, `admin_account` and `banned_account` show as error
messages. A failed native submission returns to
`/delete-account?error=…&form=delete-account`; the typed address never enters
a URL.

`/deleted` is static and reads no account: the account is deleted, removing
its data finishes shortly and a notice went to the old address. It links to the
Tadoku home page. The deletion it confirms has already been accepted, so the
root loader skips the session read on this page and never waits for the
identity API.

## Session

`frontend/apps/authv2/src/lib/session.server.ts` requests `GET /identity/session`
for every page except `/deleted` through the [generated client](api-client.md)
from the configured Tadoku API root,
forwarding only the incoming Cookie and `X-Dev-Branch` request headers. The
generated session schema checks the response shape; the
API owns account field constraints, so imported display names and email formats
pass. Session data is request-scoped; both the upstream read and rendered response disable caching.
The server function in `frontend/apps/authv2/src/lib/session.ts` keeps server
configuration and request credentials out of the browser bundle.

A missing endpoint, non-success response, timeout or invalid payload shows an
unavailable page with a retry link. An unavailable answer does not invent an
authenticated account.

`frontend/apps/authv2/src/lib/urls.server.ts` resolves origins per request.
Authv2 is the account app: development serves it on `account.tadoku.dev.lab`
and `<route>.account.tadoku.dev.lab`, production on `account.tadoku.app` and
`<route>.account.preview.tadoku.app`. The former `authv2.` hosts are a temporary
redirect: `authv2.tadoku.app` redirects to `account.tadoku.app`, keeping the
path and query, and `urls.server.ts` still accepts `authv2.` branch hosts until
that redirect is removed. A valid branch label selects the sibling Webv2 origin
and API gateway. Other hosts retain the configured origins. Forwarded host
headers cannot select an upstream.

## Local development

Use Node 22.12 or newer and pnpm. From `frontend/`:

```sh
pnpm install --frozen-lockfile
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm authv2
```

`predev` builds `paper-ui`; Vite listens on port 5184. Check existing listener
ownership before starting another instance.
Nitro's development runner uses `self` so SSR starts in the Vite process and
page requests do not race worker initialization. Source edits update the browser
through Vite's live connection.

| Variable | Default | Purpose |
| --- | --- | --- |
| `TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Webv2 origin and trusted sibling-host root |
| `TADOKU_API_URL` | `https://tadoku.dev.lab/api/internal` | Server-only API root |
| `HOST` | `0.0.0.0` | Listener address |
| `PORT` | `5184` | Vite listener port |

## Branch overlays and base

Use [Development environment](../develop/environment.md) to start a branch and
obtain its authv2 link. Authv2 source changes select its deployable; shared
`paper-ui` changes select both authv2 and Webv3. The Paper runtime uses Node 22,
builds shared declarations initially, watches Paper assets and runs the selected
app's Vite server on port 3000. Public Paper type changes require another full
`paper-ui` build. API and tenant provisioning remain separate responsibilities.

The base lives in `k8s/dev/base/frontend-authv2/`, with a native Nitro server,
Lab CA trust, API-independent `/healthz` probes and canonical gateway routes
that strip client-supplied branch headers. The development base becomes
available only after its manifests and compatible image have been deployed.
Public `/api/internal/identity/*` requests use Oathkeeper's anonymous
authenticator and signed `id_token`, retaining the original Cookie for the
identity API. They do not invoke the global session callback: the identity API
owns session authentication and CSRF, including session reads and logout for
banned users. Other API rules retain their existing authentication behavior.
Both development account apps use `tdk-dev-frontend-auth`, with separate
Deployment/Service names and selectors. Authv2 overlays set `TADOKU_WEB_URL`,
`TADOKU_API_URL` and `NODE_EXTRA_CA_CERTS` inline and mount the existing
`tadoku-cli-ca`; they do not need the base's `tadoku-cli-authv2` ConfigMap or
auth's Next.js configuration. No new shared namespace or ConfigMap is required
to verify a branch. Production retains its separate authv2 namespace.

## Checks and manual browser smoke

From `frontend/`:

```sh
pnpm --filter authv2 exec tsc --noEmit
pnpm --filter authv2 lint
pnpm --filter authv2 build
pnpm check:paper-boundaries
AUTHV2_BASE_URL=http://127.0.0.1:5184 pnpm --filter authv2 test:e2e
```

The Playwright smoke suite is run by hand, never in CI. Use an owned tenant on
v2 with native Dev Reader/Admin fixtures and the real API, gateway and worker.
`credentials.spec.ts` covers JavaScript-disabled registration and login,
delayed hydration, the browser-to-API transport, `return_to`, wrong passwords,
throttling, field errors and mobile containment. `settings.spec.ts` covers the
guest continuation, profile and email persistence, the old-address notice,
password errors, foreign-origin posts, logout and, when an operator ages the
session, stale-session reauthentication. `settings-deletion.spec.ts` covers the
Settings link as the only entry point, deletion with and without JavaScript, a
mistyped address, administrator refusal, mobile containment and, when an
operator ages the session, the fresh-login redirect and banner.
`recovery.spec.ts` follows a held recovery email's generated link through
Settings and checks old passwords, sessions and links fail. See `frontend/apps/authv2/e2e/README.md` for setup and
for the loopback session fixture that checks header forwarding and failure
responses, for the registration fixture that sends controlled
API answers, including the registration `429`, through both the hydrated and
native registration paths, and for the Settings fixtures that serve the API's
exact session bytes for imported addresses with control characters and check
every Settings URL state with and without JavaScript. These fixtures prove application HTTP/SSR behavior; it
does not prove real authentication, Oathkeeper, tenant isolation or
shared-cluster routing. TLS verification remains enabled. Reports, screenshots
and recordings go outside source control. Tracing is disabled because
authenticated network payloads contain credentials and links.

The development-base E2E always builds authv2 from the tested revision and selects
that local image through `TADOKU_BASE_AUTHV2_IMAGE`. This deliberate source check
works before or after publication; other images remain published inputs. It does
not verify authv2 publication, registry visibility or Image Updater.

## Images and deployment

`frontend/apps/authv2/Dockerfile` builds the standalone Node 22 Nitro runtime.
It runs as UID/GID 1000 on port 3000 with a read-only root filesystem and no
Kubernetes API token. `/healthz` remains healthy when the identity API is absent.
The production image defaults runtime origins to `https://tadoku.app` and
`https://tadoku.app/api/internal`; development overrides them through ConfigMaps.

`frontend/scripts/publish-image.sh` and `//frontend:authv2_branch_push` publish
explicitly requested tags from committed source. The authv2 build workflow
checks types, lint, boundaries and the container without running Playwright.
Main publication creates commit and `latest` tags; only an owner-approved manual
workflow dispatch with `publish_production` creates `prod`. Production serves
authv2 at `account.tadoku.app`; the maintainer owns its manifests, DNS,
certificates and deployment in the separate GitOps repository.

The first source merge automatically starts both image publication and the
development base's sync from `main`. The package does not exist before its
first publication; public repository permissions do not make a GHCR package
public. Follow this order:

1. Merge reviewed source with its current-head container journeys passing.
   The normal `build-frontend-authv2` workflow checks the image, then publishes
   `ghcr.io/tadoku/tadoku/frontend-authv2:<merge-sha>` and `:latest`.
2. After that publication succeeds, a package administrator opens
   [authv2 package settings](https://github.com/orgs/tadoku/packages/container/tadoku%2Ffrontend-authv2/settings)
   and changes only this new package's visibility to Public. Verify an anonymous
   pull. Publishing and changing visibility are separate actions; no manual
   workflow dispatch or early image upload is needed.
3. Require authv2's Pod Ready, `tadoku-dev` Synced and Healthy and
   `https://account.tadoku.dev.lab/login` returning 200. Authv2 shows its
   unavailable page until the identity API is deployed.
4. Enroll authv2 in the homelab Image Updater with the digest strategy and
   write-back to `k8s/dev/base/kustomization.yaml`, then verify an immutable digest
   resolution. The homelab infrastructure repository owns this configuration.

Before publication or while the package remains private, the new authv2 Pod may
stay in `ImagePullBackOff`. Its wave-30 health gate can prevent wave-50 Ingress
creation and leave the development Application's sync incomplete or failed;
existing workloads remain available. Kubelet retries the same image pull after
visibility changes. Argo recovery must be observed: the live Application has
automated sync and self-heal but no explicit `syncPolicy.retry`; a previously
observed operation carried a finite retry limit of five. Do not assume unlimited
retry or authorize a manual sync from this guide. If it remains failed after the
Pod recovers, the maintainer must assess the failed operation and approve any
required full sync separately.

Source changes to Webv3 or the workspace lockfile also trigger its normal
publication workflow. Production already follows `frontend-webv3:latest` by
digest: Image Updater writes to the production GitOps repository and Argo CD
automatically rolls `next.tadoku.app`. This is a source-merge effect. Webv3's SSR bundles
`@tanstack/query-core` to keep its build compatible with this workspace; the
separate Nitro `devServer.runner: 'self'` setting affects development only.
Keep `webv3 / container journeys` green on the exact source head before landing.

# Accounts and sessions

## Find it

Main-site **Log in** or **Sign up** opens the account host. Account routes are
`/login`, `/register`, `/account-recovery`, `/verification`, and `/` for Settings.
The logged-in main-site user menu offers Profile, Account settings and Log out;
Admin appears for administrators. Use a separate browser context for each identity.

## Verify

- Follow Log in from a main-site deep link, sign in with a synthetic fixture,
  and verify the return URL, logged-in navigation and an authenticated API result.
  Reload: session and role-dependent controls should remain correct. Account and
  main links must retain the same branch hostname.
- Open `/logs/new` and a branch contest's registration page anonymously with
  multiple query parameters and a fragment. Their login URL must carry the full
  original URL in `return_to`; after fixture login and reload, the browser must
  show that same protected page and URL. The auth suite's `login-return.spec.ts`
  uses an existing development reader and verifies TLS without recording traces
  or changing the identity or submitting activity/registration writes.
- For registration/settings/recovery changes, use a new uniquely named disposable
  identity. Exercise validation, successful submission, profile/password settings,
  forced reauthentication, logout and recovery mail. Confirm logout also removes
  access to protected pages; don't treat hiding the menu as authorization proof.
- For admin access changes, test administrator, ordinary reader and anonymous
  contexts through the real gateway. Check the API denial as well as the UI.

## Tools and traps

The existing [Playwright journey](../../../../../frontend/apps/auth/e2e/README.md)
documents installation, MailHog loopback forwarding, environment variables and the
optional administrator test. It creates its own identity and leaves mail/identity
data for diagnosis. Missing env skips tests. Its configured TLS bypass is not
certificate verification. It targets base hosts unless branch-host URLs are supplied.

Use MailHog only through the documented loopback port-forward; do not expose its
inbox publicly or reuse production mailboxes. Do not retain session cookies or
recovery tokens in shared evidence. Kratos/Keto are shared even with a branch API:
changing a shared identity's password, name, permissions or session affects others.
Branch host selection is independent of shared authentication, and `/kratos` is never
served by a frontend overlay. Don't mock CSRF, fabricate JWTs or bypass the UI
login when claiming a browser E2E.

## Identity entry point

Webv2 and admin read `GET /identity/session`, including as guests. Check that
`account_app` selects account or authv2 links on the same branch, and that
logout posts to `POST /identity/logout` before access to protected pages is
removed. Auth selects the same endpoint before starting a v1 flow and redirects
v2 requests to authv2 with their query. No frontend clears identity cookies
when an unrelated API returns 401.

Gateway configuration forwards Cookie and
X-Dev-Branch through the branch-aware API hop and adds the callback bearer;
verify isolated-database branches as well as ordinary tenants.

## Identity API accounts and sessions

[Identity accounts and sessions](../../../../../docs/docs/tadoku-api/identity.md)
describes the local v2 operations and the v1 session adapter. Use an owned
branch tenant and disposable account for registration, login, session lookup and
logout. Assert the tenant-specific cookie name, Secure/HttpOnly/SameSite flags,
fresh tokens on login, persistence, rejection after logout, and isolation from
base and another branch. Send a JSON content type and an allowed HTTPS Origin.
Record `account_app`; an unset `identity-v2` decision remains `v1`.

The base API and overlays supply the development cookie domain and Kratos public
URL.
Recovery and profile/password settings use the local identity API. Account
lookup/search use the separate identity `ReadService` for profile and moderation;
the `Accounts` service's Reader methods and account deletion remain `501` stubs.
Email list/detail
routes accept a verified tenant `email:read` bearer or a Keto-authorized
administrator; see [Retained email API](admin.md#retained-email-api).
The development base exposes identity and GET email paths through separate
public Oathkeeper rules. Verify that both the shared rules and the relevant
backend image are deployed; a frontend overlay does not deploy these rules.
The public
identity rule forwards cookies with a guest JWT; the development configuration
also switches other cookie-bearing routes to the identity callback and adds the
private SSR session rule. Verify their deployed revision after source merge;
the production gateway switch remains separately held. Confirm actual branch API
responses rather than treating an API port-forward as the public journey.
Provisioning seeds `identity-v2=true` automatically in the branch namespace.
Use the native Dev Admin/Reader accounts and documented development password;
verify their UUIDs/names, tenant cookie and session through the configured public
branch API. Reseeding must preserve branch-changed credentials, display names and sessions,
and continue through branch profile/content fixtures. New accounts retain the
canonical fixture user age; a missing fixture email held by another UUID must
fail atomically. Never edit canonical
flag policy or shared Kratos sessions to exercise local account behavior.
Backend E2Es use real, disposable providers and a tenant-aware Flipt fixture.

## Caller release smoke

Follow the ordered gates in
[Identity gateway configuration](../../../../../docs/docs/operations/identity-gateway.md#source-delivery-gates):
production read-only baseline, actual-merge development image provenance, isolated
native-v2 callback GET and invalid POST 400 with physical-database evidence, then
canonical browser/SSR smoke with the same retained v1 sessions. Require live
webv2/admin/auth builds from the merge commit before browser proof. Confirm
signed-in server-rendered document content before hydration as well as the real
Reader/Admin interactions; API curls and hydrated navigation alone are partial
proof. Revoke the retained pair publicly only after every smoke gate passes and
verify old-cookie 401. Keep secrets in the existing private storage states.

New production branches and `deploy`-labelled pull requests using these callers
are unsupported until the paired production identity routes, callback and authv2
are ready. Source delivery does not activate held production branch-tooling
adoption. See [Production branches](../../../../../docs/docs/develop/production-branches.md).

## Recovery and settings API journey

Use the branch main API URL printed by `dev url`, with an allowed HTTPS Origin
and JSON request bodies. Authv2 screens may be unavailable independently of the
API. In an owned v2 tenant, request recovery for its Dev Reader, read the held
`account.recovery` message as Dev Admin through the public cookie gateway,
follow `params.recovery_url` and require a `303` to that tenant's authv2
Settings root. A supplied `return_to` must survive as a query parameter for
continuation after Settings completion, never as the immediate destination.
Set a new password and log in with it. Confirm
actual overlay headers, tenant identity, a fresh secure cookie and rejection of
the old password and sessions. A second consume and a foreign tenant must fail;
unknown emails return the same empty `202`. Destinations that cannot fit into
the complete 2,048-byte Settings URL or 4,096-byte recovery link must fail before
lookup or consumption, with no truncation and the same result for known and
unknown emails. Stale sessions cannot change email
or password, while a valid stale session may still change its display name.
Email changes must notify the old address and password changes must revoke
other sessions.

A v2 administrator cookie requires deployed identity callback wiring and
tenant-local v2 administrator fixtures. The existing gateway also accepts an
already authorized v1 administrator cookie. Both paths require a valid
administrator session; opening a login page alone does not provide one.
Missing prerequisites must be reported explicitly. A bearer read, API
port-forward or disposable local gateway is separate evidence and does not
satisfy this journey. Do not create shared Kratos sessions, change shared
credentials/roles, or import unfinished wiring to bypass the prerequisite. Keep
recovery URLs and cookies out of screenshots and retained output. Remove the
owner's test data and overlay through `dev down` after verification.

## Source anchors

[Account pages](../../../../../frontend/apps/auth/pages/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[auth E2E](../../../../../frontend/apps/auth/e2e/account.spec.ts).

## Authv2 account screens

Authv2 uses `/` for Settings, `/login`, `/register` and `/recovery` on `authv2.tadoku.dev.lab` or the matching
`<route>.authv2.tadoku.dev.lab` branch. Read
[Authv2](../../../../../docs/docs/frontend/authv2.md) and
[its manual smoke instructions](../../../../../frontend/apps/authv2/e2e/README.md).

- Obtain the link with `dev url --owner <owner> --host authv2.tadoku.dev.lab '/login'`.
  Confirm the document's overlay `X-Dev-Backend` before claiming branch behavior.
  Authv2 shares `tdk-dev-frontend-auth` with auth, using distinct workload and
  Service selectors. Its Vite overlay supplies URLs inline and reuses the existing
  `tadoku-cli-ca`; no shared prerequisite creation is needed. Snapshot auth's
  UID, spec, generation, selectors and ConfigMaps before startup and compare
  after owner-scoped teardown. Require no owned resources and healthy auth.
- A v2 guest gets login, registration and recovery forms; signed-in navigation
  shows the display name and logout. Settings is at `/`, with `/settings`
  retained for existing caller links. A guest opening Settings goes through
  login and returns to the same Settings URL.
- Run the manual real-API smoke. With JavaScript disabled, register a disposable
  account, reload Settings, log out and log in again. Enter values before delayed
  hydration, then submit after hydration and verify those same values survive
  and that the browser posts JSON to the tenant's main-host
  `/api/internal/identity/*`, not to authv2's `/actions/*` fallback. Inspect
  labels, keyboard focus, pending states, field/form errors and mobile
  containment. Do not inject cookies or mock identity responses for this proof.
- Log in with an allowed `return_to` and require that exact destination; a
  foreign one must never be rendered, carried or followed.
- As the owned native Reader, save a display name and reload; exercise invalid
  names/passwords, taken email, throttling and stale-session feedback. A stale
  valid session can save its display name; email or password edits link to a
  login that returns to Settings. Email changes notify the old address.
- Request recovery through its form; use a separate native Admin context's cookie
  to read the newly held email. Follow its exact `params.recovery_url`, set a
  password in Settings, continue, log out and log in using the new password.
  Require old-password, old-session and used-link rejection. Settings offers a
  valid continuation as a link and follows it only after a password save;
  crafted off-allowlist queries are ignored. Invalid recovery links retain API
  400 JSON.
- Keep a page mounted, change source and observe HMR with unchanged navigation
  time origin, Pod UID and image. Keep the local generated route tree current:
  synchronizing a stale tree makes the overlay regenerate it and reload.
- A v1 session answer returns to the matching existing account host with the
  path and query preserved, including legacy emoji display names and email
  formats. Session API failures show an unavailable message
  with a retry action and the shell's home link; a missing endpoint must not
  crash the page.
- Verify mobile containment and `/healthz` independently of API availability.
  Cookie/branch forwarding can be checked with the opt-in local session fixture;
  disclose that substituted upstream instead of claiming real login coverage.
  The Playwright smoke is run by hand, never wired into CI.
- Public identity entry rules forward Cookie with a signed guest JWT and no
  global session callback, so identity owns auth/CSRF and banned-user session or
  logout handling. The base's disposable Oathkeeper E2E proves forwarding and
  unchanged unrelated-route denial using synthetic HTTP fixtures. Until the
  changed shared Oathkeeper and identity backend are deployed, that combined
  live boundary remains unverified; a frontend overlay does not replace it.

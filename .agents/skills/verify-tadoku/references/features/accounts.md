# Accounts and sessions

## Find it

Main-site **Log in** or **Sign up** opens authv2 on the matching account host. Its
routes are `/login`, `/register`, `/recovery`, and `/` or `/settings` for Settings.
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
  show that same protected page and URL.
- For registration/settings/recovery changes, use a new uniquely named disposable
  identity. Exercise validation, successful submission, profile/password settings,
  forced reauthentication, logout and recovery mail. Confirm logout also removes
  access to protected pages; don't treat hiding the menu as authorization proof.
- For admin access changes, test administrator, ordinary reader and anonymous
  contexts through the real gateway. Check the API denial as well as the UI.

## Tools and traps

The [authv2 Playwright smoke](../../../../../frontend/apps/authv2/e2e/README.md)
documents installation, environment variables and its opt-in local fixtures; see
[Authv2 account screens](#authv2-account-screens).

Do not retain session cookies or recovery tokens in shared evidence. Canonical
accounts serve everyone using the base: changing their password, name, roles or
session affects others. Roles belong to the tenant that stores them.
Don't mock CSRF, fabricate JWTs or bypass the UI login when claiming a browser E2E.

## Identity entry point

Webv2 and admin read `GET /identity/session`, including as guests. Check that
login, registration and settings links open authv2 on the same branch, and that
logout posts to `POST /identity/logout` before access to protected pages is
removed. No frontend clears identity cookies when an unrelated API returns 401.

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
Record `account_app`. Only an `identity-v2` flag absent from the canonical
`tadoku/prod` snapshot defaults to `v1`. Any other undecided selection returns
`503`, or `504` past the request deadline and `499` when the client cancels.

The base API and overlays supply the development cookie domain.
Recovery and profile/password settings use the local identity API. Account
lookup/search use the separate identity `ReadService` for profile and moderation;
the `Accounts` service's Reader methods and account deletion remain `501` stubs.
Email list/detail
routes accept a verified tenant `email:read` bearer or an administrator of
that tenant; see [Retained email API](admin.md#retained-email-api).
The development base exposes identity and GET email paths through separate
public Oathkeeper rules. Verify that both the shared rules and the relevant
backend image are deployed; a frontend overlay does not deploy these rules.
The public
identity rule forwards cookies with a guest JWT; the development configuration
also switches other cookie-bearing routes to the identity callback and adds the
private SSR session rule. Verify their deployed revision after source merge.
Confirm actual branch API responses rather than treating an API port-forward as
the public journey.
Provisioning seeds `identity-v2=true` automatically in the branch namespace.
Use the native Dev Admin/Reader accounts and documented development password;
verify their UUIDs/names, tenant cookie and session through the configured public
branch API. Reseeding must preserve branch-changed credentials, display names and sessions,
and continue through branch profile/content fixtures. New accounts retain the
canonical fixture user age; a missing fixture email held by another UUID must
fail atomically. Never edit canonical
flag policy or canonical sessions to exercise local account behavior.
Backend E2Es use real, disposable providers and a tenant-aware Flipt fixture.

## Caller release smoke

Follow the ordered gates in
[Identity gateway configuration](../../../../../docs/docs/operations/identity-gateway.md#source-delivery-gates):
production read-only baseline, actual-merge development image provenance, isolated
native-v2 callback GET and invalid POST 400 with physical-database evidence, then
canonical browser/SSR smoke with the same retained v1 sessions. Require live
webv2/admin builds from the merge commit before browser proof. Confirm
signed-in server-rendered document content before hydration as well as the real
Reader/Admin interactions; API curls and hydrated navigation alone are partial
proof. Revoke the retained pair publicly only after every smoke gate passes and
verify old-cookie 401. Keep secrets in the existing private storage states.

## Recovery and settings API journey

Use the branch main API URL printed by `dev url`, with an allowed HTTPS Origin
and JSON request bodies. Authv2 screens may be unavailable independently of the
API. In an owned v2 tenant, request recovery for its Dev Reader, read the held
`account.recovery` message as Dev Admin through the public cookie gateway,
follow `params.recovery_url` and require a `303` to that tenant's authv2
Settings URL `/?edit=password`, which opens the Password editor. A supplied
`return_to` must follow it as `&return_to=` for continuation after Settings
completion, never as the immediate destination.
Set a new password and log in with it. Confirm
actual overlay headers, tenant identity, a fresh secure cookie and rejection of
the old password and sessions. A second consume and a foreign tenant must fail;
unknown emails return the same empty `202`. Within 15 minutes, one address
receives at most three recovery emails, with later requests returning that
`202` without mail, and one client IP gets ten requests before `429`. Destinations that cannot fit into
the complete 2,048-byte Settings URL or 4,096-byte recovery link must fail before
lookup or consumption, with no truncation and the same result for known and
unknown emails. Stale sessions cannot change email
or password, while a valid stale session may still change its display name.
Email changes must notify the old address and password changes must revoke
other sessions. Stored addresses receive mail without the spaces and tabs around them; one
the email writer still rejects makes recovery and email changes return `400`
without writes.

A v2 administrator cookie requires deployed identity callback wiring and
tenant-local v2 administrator fixtures. It requires a valid administrator
session; opening a login page alone does not provide one.
Missing prerequisites must be reported explicitly. A bearer read, API
port-forward or disposable local gateway is separate evidence and does not
satisfy this journey. Do not change shared credentials/roles or import
unfinished wiring to bypass the prerequisite. Keep
recovery URLs and cookies out of screenshots and retained output. Remove the
owner's test data and overlay through `dev down` after verification.

## Source anchors

[Account screens](../../../../../frontend/apps/authv2/src/routes/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[authv2 E2E](../../../../../frontend/apps/authv2/e2e/README.md).

## Authv2 account screens

Authv2 uses `/` for Settings, `/login`, `/register` and `/recovery` on `account.tadoku.dev.lab` or the matching
`<route>.account.tadoku.dev.lab` branch. Read
[Authv2](../../../../../docs/docs/frontend/authv2.md) and
[its manual smoke instructions](../../../../../frontend/apps/authv2/e2e/README.md).

- Obtain the link with `dev url --owner <owner> --host account.tadoku.dev.lab '/login'`.
  Confirm the document's overlay `X-Dev-Backend` before claiming branch behavior.
  Authv2 runs in `tdk-dev-frontend-auth`. Its Vite overlay supplies URLs inline
  and reuses the existing `tadoku-cli-ca`; no shared prerequisite creation is
  needed. After owner-scoped teardown, require no owned resources and a healthy
  base authv2.
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
- Settings rows open one editor at a time through their **Change** links, which
  set `?edit=`; Change and Cancel must work without JavaScript and keep
  `return_to`. A save collapses the row and shows its confirmation.
- As the owned native Reader, save a display name and reload; exercise invalid
  names/passwords, taken email, throttling and stale-session feedback. A stale
  valid session can save its display name; email or password edits link to a
  login that returns to Settings with that editor open. Email changes notify the
  old address.
  Resubmitting the unchanged email shows that it is already the account's
  address and sends no request; display-name and password saves send only their
  own field.
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
- Session API failures show an unavailable message with a retry action and
  the shell's home link; a missing endpoint must not crash the page.
- Verify mobile containment and `/healthz` independently of API availability.
  Cookie/branch forwarding can be checked with the opt-in local session fixture,
  registration answers, including the registration budget's `429`, with the
  opt-in registration fixture, and Settings for imported addresses with control
  characters with the opt-in Settings fixture; disclose that substituted upstream
  instead of claiming real login coverage.
  The Playwright smoke is run by hand, never wired into CI.
- Public identity entry rules forward Cookie with a signed guest JWT and no
  global session callback, so identity owns auth/CSRF and banned-user session or
  logout handling. The base's disposable Oathkeeper E2E proves forwarding and
  unchanged unrelated-route denial using synthetic HTTP fixtures. Until the
  changed shared Oathkeeper and identity backend are deployed, that combined
  live boundary remains unverified; a frontend overlay does not replace it.

## Webv3 account bridge

Use `dev url --owner <owner> --host next.tadoku.dev.lab '/plan?source=next#entry'`.
Anonymous HTML shows Log in with path and search; hydration adds the fragment.
Sign in through authv2 on the sibling account host, confirm the exact return and
display name, reload, then POST Log out and verify home is anonymous, the
tenant's `tadoku_session_<tenant>` cookie is cleared and `GET /identity/session`
on the branch API returns no account for the old cookie. Check a foreign-origin
logout returns 403 without ending the session.

Use Reader, Admin and anonymous contexts for thirty interleaved SSR reads. Each
response must show only its viewer and disable caching. The application emits
private, no-store; development ingress replaces it with no-store. Keep traces,
videos and storage dumps off for authenticated runs. Production login returns
and logout also need `next.tadoku.app` and `<route>.next.preview.tadoku.app` on
the identity API's app host allowlist and in authv2's continuation check.

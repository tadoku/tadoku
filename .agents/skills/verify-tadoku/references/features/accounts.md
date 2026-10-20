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
Account lookup/search use the separate identity `ReadService` for profile and
moderation; the `Accounts` service's Reader methods, recovery/settings and account
deletion remain `501` stubs. Email list/detail
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
Branch v2 flag provisioning must use an authorized owner/tenant mechanism. A temporary, uncommitted `identity-v2` seed
in the checkout before `dev up` is confined to that owner's test namespace;
restore it after teardown and never apply it to canonical Flipt. Never edit
canonical flag policy or shared Kratos sessions to exercise local account
behavior. Backend E2Es use real, disposable providers and a tenant-aware Flipt fixture.

## Source anchors

[Account pages](../../../../../frontend/apps/auth/pages/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[auth E2E](../../../../../frontend/apps/auth/e2e/account.spec.ts).

## Authv2 skeleton

Authv2 uses `/` and `/login` on `authv2.tadoku.dev.lab` or the matching
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
- A tenant selecting v2 shows the Paper skeleton and anonymous/signed-in state;
  registration, recovery and settings are not available in this foundation.
- A v1 session answer returns to the matching existing account host with the
  path and query preserved, including legacy emoji display names and email
  formats. Session API failures show an unavailable message
  and existing-account link; a missing endpoint must not crash the page.
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

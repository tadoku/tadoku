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

## Identity API skeleton

[Identity contracts](../../../../../docs/docs/tadoku-api/identity.md) lists the
new API and callback routes. They return `501` with the shared JSON error after
their route group's authentication. Missing JWT credentials return `400`, invalid
JWTs `401`, and missing/wrong callback credentials `401` before the stub.
Email-only token verification is also a `501` stub and admits no reader.

Identity paths are exposed once the dev base deploys its public identity rule;
email paths are not yet exposed by Oathkeeper's public rules. Before that base
deployment, record the missing gateway exposure explicitly when verifying a
branch; a frontend overlay does not deploy this shared rule. A read-only API
port-forward plus a real Oathkeeper decisions JWT can verify server routes, but
does not prove gateway exposure or a browser login journey. Verify an existing
public endpoint's `X-Dev-Backend` separately. An unset `identity-v2` flag must
evaluate false for the branch tenant; do not change shared flag policy to exercise it.

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

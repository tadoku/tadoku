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

## Identity API accounts and sessions

[Identity accounts and sessions](../../../../../docs/docs/tadoku-api/identity.md)
describes the local v2 operations and the v1 session adapter. Use an owned
branch tenant and disposable account for registration, login, session lookup and
logout. Assert the tenant-specific cookie name, Secure/HttpOnly/SameSite flags,
fresh tokens on login, persistence, rejection after logout, and isolation from
base and another branch. Send a JSON content type and an allowed HTTPS Origin.
Record `account_app`; an unset `identity-v2` decision remains `v1`.

The API overlay supplies its development cookie domain and Kratos public URL.
Recovery/settings, email/token routes and account lookup/search are still
skeletons. Gateway exposure of `/identity` and the Oathkeeper callback switch are
separate prerequisites: record a gateway 404 explicitly instead of treating an
API port-forward as the public journey. Branch v2 flag provisioning must use an
authorized owner/tenant mechanism; never edit canonical flag policy or shared
Kratos sessions to exercise local account behavior. Backend E2Es use real,
disposable providers and a tenant-aware Flipt fixture.

## Source anchors

[Account pages](../../../../../frontend/apps/auth/pages/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[auth E2E](../../../../../frontend/apps/auth/e2e/account.spec.ts).

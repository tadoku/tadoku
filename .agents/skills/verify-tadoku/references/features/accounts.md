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

The callback and session/logout runtime must be integrated before claiming this
journey works on a branch. Gateway configuration forwards Cookie and
X-Dev-Branch through the branch-aware API hop and adds the callback bearer;
verify isolated-database branches as well as ordinary tenants.

## Identity API skeleton

[Identity contracts](../../../../../docs/docs/tadoku-api/identity.md) lists the
new API and callback routes. They return `501` with the shared JSON error after
their route group's authentication. Missing JWT credentials return `400`, invalid
JWTs `401`, and missing/wrong callback credentials `401` before the stub.
Email-only token verification is also a `501` stub and admits no reader.

These paths are not yet exposed by Oathkeeper's public rules. Record that
boundary explicitly when verifying a branch. A read-only API port-forward plus
a real Oathkeeper decisions JWT can verify server routes, but does not prove the
missing gateway exposure or a browser login journey. Verify an existing public
endpoint's `X-Dev-Backend` separately. An unset `identity-v2` flag must evaluate
false for the branch tenant; do not change shared flag policy to exercise it.

## Source anchors

[Account pages](../../../../../frontend/apps/auth/pages/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[auth E2E](../../../../../frontend/apps/auth/e2e/account.spec.ts).

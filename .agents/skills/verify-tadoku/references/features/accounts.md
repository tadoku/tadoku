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

## Source anchors

[Account pages](../../../../../frontend/apps/auth/pages/),
[session navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[auth E2E](../../../../../frontend/apps/auth/e2e/account.spec.ts).

## Webv3 account bridge

Use `dev url --owner <owner> --host next.tadoku.dev.lab '/plan?source=next#entry'`.
Anonymous HTML shows Log in with path and search; hydration adds the fragment.
Sign in through the sibling account host, confirm the exact return and display
name, reload, then POST Log out and verify home is anonymous and the old session
is invalid. Check a foreign-origin logout returns 403 without ending the session.

Use Reader, Admin and anonymous contexts for thirty interleaved SSR reads. Each
response must show only its viewer and disable caching. The application emits
private, no-store; development ingress replaces it with no-store. Keep traces,
videos and storage dumps off for authenticated runs. Canonical production and
branch preview login-return acceptance are separate release gates.

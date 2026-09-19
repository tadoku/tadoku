# Kratos browser checks

Run against the Tilt development stack with Kratos and MailHog ready. The account
journey creates a unique `@example.com` test identity, changes its password and
profile, and sends recovery mail to MailHog. It does not reset the database or
modify existing accounts. Test identities and mail remain available for debugging.
Do not point this test at production.

```sh
cd frontend
pnpm --filter auth exec playwright install --with-deps chromium
AUTH_URL=https://account.example.test \
APP_URL=https://app.example.test \
MAILHOG_URL=https://app.example.test/mail \
pnpm --filter auth test:e2e
```

Use the hosts from your ignored Tilt configuration. The checks exercise the real
browser UI, Kratos cookies/CSRF, server-rendered sessions, settings, forced
reauthentication, logout, and the recovery link sent by the courier. Screenshots
from failed checks are kept in the ignored `test-results/` directory.

To include the existing-admin journey, also set `ADMIN_URL`, `E2E_ADMIN_EMAIL`
and `E2E_ADMIN_PASSWORD` to a seeded development administrator. It checks the
login return URL, the admin dashboard after reload, the real authorization API,
and logout from the account site. Missing environment variables cause the
corresponding journey to be reported as skipped.

The dev Kratos configuration explicitly uses registration `style: unified` to
keep the combined profile/password form on v26.2.0. Production must select the
same registration style when this frontend is deployed there.

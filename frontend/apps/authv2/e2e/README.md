# Authv2 manual smoke

Run from `frontend/` with Node 22.12+, pnpm and a running authv2 server. This
Playwright smoke suite is deliberately not wired into CI.

```sh
pnpm --filter authv2 exec playwright install chromium
AUTHV2_BASE_URL=https://<route>.account.tadoku.dev.lab \
  AUTHV2_READER_STATE_FILE=<private file outside the report> \
  PLAYWRIGHT_OUTPUT_DIR=/workspace/.scratch/authv2-smoke \
  pnpm --filter authv2 exec playwright test e2e/credentials.spec.ts e2e/recovery.spec.ts e2e/settings.spec.ts
```

The tenant must provide native Dev Reader and Dev Admin accounts, with the
documented development passwords. The suite writes only to that owned
branch tenant and reads held email through the tenant's main host with an Admin
cookie. The recovery journey replaces the branch Reader password and saves the
replacement with mode 0600 in `AUTHV2_READER_STATE_FILE`, so repeat runs can log
in; delete the file after owner teardown. Optional `AUTHV2_READER_PASSWORD` and
`AUTHV2_ADMIN_PASSWORD` select private fixture credentials.

A full run makes about ten Reader logins and eleven throttled attempts for one
random address. The identity API allows ten attempts per email and fifty per
address in fifteen minutes, so space repeat runs or reset only the owned
tenant's `tenant:tadoku/<route>:identity:login:*` counters between them.

The stale-session test is skipped unless `AUTHV2_STALE_SESSION_DIR` names a
private directory. It logs the Reader in, writes `ready` there and waits for an
`aged` file. Move only that tenant's Reader sessions' `authenticated_at` back
more than an hour with an owner-scoped task, then create `aged`.

TLS verification stays enabled. Reports and videos contain synthetic fixture
screens only. Tracing is disabled to avoid retaining authenticated requests or
recovery URLs. Inspect and sanitize all artifacts before publication. The
optional local fixture below tests session transport only; run
`pnpm --filter authv2 exec playwright test e2e/smoke.spec.ts` for that fixture,
never the real account journey.

For deterministic local HTTP/SSR coverage, first check ports 5184 and 5195 are
free and start this checkout's server in a separate terminal:

```sh
TADOKU_WEB_URL=http://localhost:5195 \
TADOKU_API_URL=http://127.0.0.1:5195/api/internal \
HOST=127.0.0.1 pnpm authv2
```

Then run:

```sh
AUTHV2_SESSION_FIXTURE=true \
  PLAYWRIGHT_OUTPUT_DIR=/workspace/.scratch/authv2-smoke \
  pnpm --filter authv2 test:e2e
```

The opt-in test fixture binds only `127.0.0.1:5195` and closes after the test.
It returns synthetic session responses and records synthetic request headers.
Assertions prove the application's server forwards Cookie and `X-Dev-Branch`,
and 404/503/malformed responses show an unavailable message with a retry action
and the shell's home link. A second test renders Settings for stored account
values: a version-0 account ID, and imported emails with an ideographic-space
prefix or 321 raw bytes. Each must appear unchanged in the server HTML and the
hydrated input and fit 1280px and 390px viewports. Resubmitting the unchanged
address must send no `PATCH /identity/settings/profile`; a value typed before
hydration must be the one sent.
This substitutes only the identity API; real authentication, gateway routing,
DNS, TLS termination and tenant persistence remain unverified. Never point
this fixture at a shared API or treat it as live end-to-end authentication.

The registration fixture needs a server whose web URL is HTTPS, because native
submissions are accepted only from an HTTPS allowed account origin. Stop the
session fixture's server, check the same ports are free, and start:

```sh
TADOKU_WEB_URL=https://tadoku.dev.lab \
TADOKU_API_URL=http://127.0.0.1:5195/api/internal \
HOST=127.0.0.1 pnpm authv2
```

Then run:

```sh
AUTHV2_REGISTRATION_FIXTURE=true \
  PLAYWRIGHT_OUTPUT_DIR=/workspace/.scratch/authv2-smoke \
  pnpm --filter authv2 exec playwright test e2e/registration.spec.ts
```

It answers registration with a `201` session, the registration budget's `429`,
`409`, `400`, `503`, the undocumented `504` and `499`, an HTML `429`, a
malformed `201`, an undocumented `200` and a reset connection. The hydrated test
answers the browser's calls to `https://tadoku.dev.lab` inside Playwright, so
they never reach the network; the native test posts to `/actions/register` as
`account.tadoku.dev.lab`, and the server forwards to the `127.0.0.1:5195`
fixture. Each failure must keep the visitor on the form with the expected
message after exactly one request, and only the `201` session may continue or
pass a session cookie on.

The Settings fixture uses the same server. Run:

```sh
AUTHV2_SETTINGS_FIXTURE=true \
  PLAYWRIGHT_OUTPUT_DIR=/workspace/.scratch/authv2-smoke \
  pnpm --filter authv2 exec playwright test e2e/settings-email.spec.ts
```

It serves the session bytes the API writes for an ordinary address and for
imported addresses with a leading vertical tab, form feed or CRLF, or a
trailing LF. The email input shows the address without line breaks or
surrounding ASCII whitespace. With and without JavaScript, resubmitting the
unchanged address must send nothing and show that it is already the account's
address; display-name and password saves must send only their own field, and an
edited address must be sent as typed. The JavaScript-disabled test captures each
native form body the browser posts, then replays it to `/actions/*` as
`account.tadoku.dev.lab`.

`AUTHV2_BASE_URL` defaults to `http://127.0.0.1:5184`.
`PLAYWRIGHT_OUTPUT_DIR` defaults to a temporary directory. Reports and mobile
screenshots remain outside Git; tracing stays off because network payloads can contain credentials. Inspect evidence before attaching it to a PR. Stop only the
server you started when verification is finished.

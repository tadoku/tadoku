# Authv2 manual smoke

Run from `frontend/` with Node 22.12+, pnpm and a running authv2 server. This
single Playwright test is deliberately not wired into CI.

```sh
pnpm --filter authv2 exec playwright install chromium
AUTHV2_BASE_URL=https://<route>.authv2.tadoku.dev.lab \
  PLAYWRIGHT_OUTPUT_DIR=/workspace/.scratch/authv2-smoke \
  pnpm --filter authv2 test:e2e
```

The tenant must select v2. The browser checks the visible skeleton and mobile
containment; the HTTP client checks API-independent health. TLS verification
stays enabled. No account or provider data is written.

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
v1 returns HTTP 307 to the existing account app preserving path/query (including
legacy accounts with 25 emoji in their display name or a local-domain email), and
404/503/malformed responses keep an unavailable message and account link.
This substitutes only the identity API; real authentication, gateway routing,
DNS, TLS termination and tenant persistence remain unverified. Never point
this fixture at a shared API or treat it as live end-to-end authentication.

`AUTHV2_BASE_URL` defaults to `http://127.0.0.1:5184`.
`PLAYWRIGHT_OUTPUT_DIR` defaults to a temporary directory. Reports and mobile
screenshots remain outside Git; failed tests retain a trace, which can contain
synthetic cookies. Inspect evidence before attaching it to a PR. Stop only the
server you started when verification is finished.

# Identity session resilience

These browser checks render the three actual Next applications against an
explicit API fixture. They prove outage, fail-closed UI and retry behavior;
they do not prove login, gateway authentication, provider sessions or tenancy.

Install the locked frontend dependencies and Playwright Chromium. Reserve six
unused loopback ports (43270–43275 here), check for an existing server belonging
to this exact checkout, and use an evidence directory outside the repository.
Run these commands from the repository root in separate owned terminals:

```sh
export IDENTITY_UI_STATE=/workspace/.scratch/identity-resilience/ui-state.json
export IDENTITY_UI_EVIDENCE=/workspace/.scratch/identity-resilience/report
mkdir -p "$(dirname "$IDENTITY_UI_STATE")" "$IDENTITY_UI_EVIDENCE"
printf '{"status":503}' > "$IDENTITY_UI_STATE"
node frontend/apps/auth/verification/fixture.cjs

NEXT_SERVER_API_ENDPOINT=http://127.0.0.1:43270/api/internal \
NEXT_PUBLIC_API_ENDPOINT=http://127.0.0.1:43270/api/internal \
pnpm -C frontend --filter webv2 exec next dev -p 43271

NEXT_SERVER_API_ENDPOINT=http://127.0.0.1:43274/api/internal \
NEXT_PUBLIC_API_ENDPOINT=http://127.0.0.1:43274/api/internal \
pnpm -C frontend --filter admin exec next dev -p 43272

NEXT_SERVER_API_ENDPOINT=http://127.0.0.1:43275/api/internal \
NEXT_PUBLIC_API_ENDPOINT=http://127.0.0.1:43275/api/internal \
pnpm -C frontend --filter auth exec next dev -p 43273

IDENTITY_WEB_ORIGIN=http://127.0.0.1:43270 \
IDENTITY_ADMIN_ORIGIN=http://127.0.0.1:43274 \
IDENTITY_AUTH_ORIGIN=http://127.0.0.1:43275 \
pnpm -C frontend --filter auth exec playwright test \
  --config verification/identity-resilience.config.ts
```

The test config accepts `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` for an installed
browser and `PLAYWRIGHT_LAUNCH_ARGS` for host-specific launch requirements. The
report retains JSON results, screenshots and failure traces. Record the tested
Git revision and diff beside it. Stop only the four processes started for this
run. No Docker, live providers, shared identities or cluster setup are involved.

## Built standalone server

After building webv2, admin and auth, reserve unused loopback port 43276 and run:

```sh
node frontend/apps/auth/verification/standalone.cjs \
  /workspace/.scratch/identity-resilience/standalone
```

The check starts each actual built standalone server in turn with an API runtime
override. A preload blocks external fetches, and a loopback identity fixture
records the complete synthetic cookie header and branch header. Webv2/admin
must render without a session outage; auth must select v2 before any legacy
provider flow. JSON, rendered HTML and startup output are retained; each child
server is stopped. This proves runtime URL selection and header forwarding,
not real credentials, DNS, gateway routing or Kratos behavior.

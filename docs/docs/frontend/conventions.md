---
title: Frontend conventions
description: The binding rules for frontend code in Tadoku, covering pnpm, the legacy ui design system, react-hook-form, API data parsing, Paper boundaries and the checks to run.
---

# Frontend conventions

Read this before you change code in `frontend/`.

Apply [Module design](../architecture/module-design.md) to components and hooks.
Hide interaction mechanics in controls and request, parsing and cache mechanics
in data hooks; screens own product state and effects.

## Tooling

Always use `pnpm`, never `npm`. Run commands from `frontend/`.

## Legacy applications

webv2, auth, admin and styleguide use the `ui` package. Never write custom
button or form styles:

- Buttons use `className="btn"` with the variants `primary`, `secondary`,
  `danger` and `ghost`.
- Form inputs such as `Input`, `Select` and `Checkbox` come from
  `ui/components/Form`.
- Shared components such as `Modal`, `Flash` and `Navbar` come from `ui`.

## Forms

Always use `react-hook-form`; never keep form fields in plain `useState`.

- Set up form context with `useForm()` and `<FormProvider>`.
- Use `<Input>`, `<Select>` and `<TextArea>` from `ui`; they read the form
  context through `useFormContext()`.
- Wrap custom or non-standard inputs, such as the code editor, with
  `useController()`.
- Submit with `methods.handleSubmit()` so built-in validation runs.
- Use `methods.watch()` for reactive values such as live previews, and
  `methods.reset()` to load existing data.

## Cross-app URLs

Use `useAppUrls()`, `browserAppUrls()` or `appUrlsForHost()` from
`frontend/packages/ui/app-urls.tsx` for home, account, admin and browser API
URLs. Resolve them per request or browser call so branch hosts retain their
sibling origins. Supply `AppUrlsProvider` from each application's request
properties for consistent SSR and hydration. Kratos stays on the base account
host. Avoid module-level runtime URL constants.

Production branch hosts use `<branch>.preview.tadoku.app`,
`<branch>.account.preview.tadoku.app` and
`<branch>.admin.preview.tadoku.app`. The resolver accepts one DNS label beneath
these reserved roots and preserves paths, ports and private server API endpoints.
Other production subdomains do not select a branch. Development keeps the
branch label directly beneath each configured development app host.

Server-side API calls use the private endpoint from
`serverApiEndpointForHost()` and forward the incoming `x-dev-branch` header,
which branch host routes set and production edge routes remove, so the
gateway signs the branch tenant.

Kratos builds redirects and email links from its production `ui_url`s, so the
auth application routes them through `frontend/apps/auth/src/branch.ts`:

- `followKratosUrl()` follows `redirect_browser_to` and Kratos UI anchors.
  On a branch host, `branchUrl()` moves production app URLs to the sibling
  branch host; URLs under the Kratos public endpoint stay on the base host.
- `flowReturnTo()` defaults a new flow's `return_to` to the branch account host
  on branch hosts, so Kratos remembers the branch for later email links.
- `forwardBranchFlow()` runs on every flow fetched by id. When the flow's
  `return_to`, or the `return_to` in its `request_url`, is on a branch host
  that `resolveAppUrls()` recognises, `branchFlowUrl()` sends the browser to
  the same page and query on that branch's account host. Pages already on that
  branch, production flows and foreign hosts stay where they are.

## Authentication flows

The auth application’s shared logout hook creates a Kratos logout flow only
when `useSession()` has an authenticated session. Anonymous login pages must
not create logout flows: their CSRF cookies can race the login flow. Session
changes enable logout without requiring callers to reload the page.

## API response parsing

webv2 and admin call Tadoku API with React Query hooks and parse every response
with Zod before using it ([ADR 003](../adr/003-zod.md)). The contract is
`services/tadoku-api/spec/openapi.yaml`.

## Paper applications

paper-styleguide, paper-playground and `paper-ui` follow
[Paper composition](paper-composition.md). Keep application services and
routers out of `paper-ui`. `pnpm check:paper-boundaries` rejects `paper-ui` in
legacy applications and `ui`, Next.js or Headless UI in Paper code.

## Checks

```sh
pnpm --filter <app> exec tsc --noEmit   # typecheck (Paper apps: pnpm --filter <app> typecheck)
pnpm --filter <app> lint
pnpm --filter <app> test                # when the app defines it
pnpm build                              # before creating a pull request
```

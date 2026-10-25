---
title: Frontend conventions
description: The binding rules for frontend code in Tadoku, covering pnpm, the legacy ui design system, TanStack Form and React Hook Form, calling Tadoku API, Paper boundaries and the checks to run.
---

# Frontend conventions

Read this before you change code in `frontend/`.

Apply [Module design](../architecture/module-design.md) to components and hooks.
Hide interaction mechanics in controls and request, parsing and cache mechanics
in data hooks; screens own product state and effects.

## Tooling

Always use `pnpm`, never `npm`. Run commands from `frontend/`.

## Legacy applications

webv2, admin and styleguide use the `ui` package. Never write custom
button or form styles:

- Buttons use `className="btn"` with the variants `primary`, `secondary`,
  `danger` and `ghost`.
- Form inputs such as `Input`, `Select` and `Checkbox` come from
  `ui/components/Form`.
- Shared components such as `Modal`, `Flash` and `Navbar` come from `ui`.
- Wrap every table in `.table-container`. It scrolls horizontally and shows edge
  shadows while content is off-screen in that direction. Keep the table and its
  rows transparent (translucent tints are fine) so the shadows show. When the table
  has no focusable content, give the wrapper `role="region"`, `tabIndex={0}` and
  `aria-labelledby` (or `aria-label`) so keyboard users can scroll it.

## Forms

Legacy applications use `react-hook-form`. Paper applications use TanStack Form
with `usePaperForm` from `paper-ui`. Never keep form fields in plain `useState`.
Keep legacy forms on React Hook Form without bridging the two libraries.

Paper form controls compose through
`usePaperForm`, `form.AppForm`, `form.Form` and `form.AppField`; render the bound
control from the field, such as `field.Input`. Screens own default values,
TanStack validators and submission effects. `required` marks the native control;
add the corresponding rule to the field or form validators. Number inputs emit
numbers (an empty input emits `NaN`), checkboxes own booleans or string arrays,
and autocomplete controls retain their selected objects. Disabled controls keep
their form values; screen submission handlers decide which data to persist.
Use `form.reset(values, { keepDefaultValues: true })` to load another draft
without replacing the hook's default values on the next render.
Paper consumers require TypeScript 5.4 or newer and the pinned TanStack Form peer.

For legacy applications:

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
URLs. The account URL is always the `account.` sibling of the home URL. Resolve
them per request or browser call so branch hosts retain their sibling origins.
Supply `AppUrlsProvider` from each application's request properties for
consistent SSR and hydration. Avoid module-level runtime URL constants.

Production branch hosts use `<branch>.preview.tadoku.app`,
`<branch>.account.preview.tadoku.app` and
`<branch>.admin.preview.tadoku.app`. The resolver accepts one DNS label beneath
these reserved roots and preserves paths, ports and private server API endpoints.
Other production subdomains do not select a branch. Development keeps the
branch label directly beneath each configured development app host.

Server-side API calls use the private endpoint from
`serverApiEndpointForHost()` and forward the incoming `x-dev-branch` header,
which branch host routes set and production edge routes remove, so the
gateway signs the branch tenant. The resolver reads `NEXT_SERVER_API_ENDPOINT`
at request time, ahead of the built Next.js runtime config, so standalone
production servers honor their Deployment's private endpoint.

## Authentication flows

Webv2 protected pages use `useSessionOrRedirect()` from
`frontend/apps/webv2/app/common/session.ts`. It waits for the client router and
the complete current location before sending an anonymous visitor to login.
The shared auth route helpers encode `return_to` so queries and fragments remain
part of the original destination. Registration uses the same protected-page hook.

Login, registration and settings links always open authv2. Webv2 and admin
read `GET /identity/session` on every page request, including guests, and
logout uses `POST /identity/logout`. `frontend/packages/ui/identity.ts` calls
both through the generated client; the API owns cookies. SSR forwards Cookie
and X-Dev-Branch to the private API endpoint, and responses use
`Cache-Control: private, no-store`. Session transport errors, invalid responses
and its bounded timeout produce a retry screen without authorizing protected
content. A protected query returning 401 stops without automatic reload or
retries; the visitor can retry explicitly. Logout failures show a message and
retain the current UI session until revocation succeeds. A fresh session read
on each navigation keeps session expiry current without sharing a snapshot
between tenants.

Webv3 reads `GET /identity/session` and the tenant role on the server,
forwarding only Tadoku session cookies to the request's public API gateway. Its
login link opens authv2 and upgrades the SSR path and search to the full browser
URL after hydration; without JavaScript fragments cannot be preserved. Logout is
an ordinary POST form that the server completes with `POST /identity/logout`.
Its mutation boundary requires the same Origin and same-origin Fetch Metadata,
and personalized responses disable shared caching.

## Calling Tadoku API

Every frontend calls Tadoku API through the client generated from
`services/tadoku-api/spec/openapi.yaml`, whose operations validate responses with
Zod ([ADR 003](../adr/003-zod.md)). Never write an API URL, request or response
type, or response schema by hand; see [Tadoku API client](api-client.md).

## Paper applications

webv3, paper-styleguide, paper-playground and `paper-ui` follow
[Paper composition](paper-composition.md). Keep application services and
routers out of `paper-ui`. `pnpm check:paper-boundaries` rejects `paper-ui` in
legacy applications and `ui`, Next.js or Headless UI in Paper code.

Webv3 uses TanStack Start, Router and Query with request-scoped server rendering,
Zod response validation and TanStack Form. Its data source and local development
commands are documented in [Webv3](webv3.md).

## Checks

```sh
pnpm --filter <app> exec tsc --noEmit   # typecheck (Paper apps: pnpm --filter <app> typecheck)
pnpm --filter <app> lint
pnpm --filter <app> test                # when the app defines it
pnpm build                              # before creating a pull request
```

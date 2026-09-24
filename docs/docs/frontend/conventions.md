---
title: Frontend conventions
description: The binding rules for frontend code in Tadoku, covering pnpm, the legacy ui design system, react-hook-form, API data parsing, Paper boundaries and the checks to run.
---

# Frontend conventions

Read this before you change code in `frontend/`.

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

## API data

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

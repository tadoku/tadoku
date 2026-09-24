---
title: Verifying changes
description: How to prove a Tadoku change works in the development environment, which checks to run before a pull request, and what evidence to attach to it.
sidebar_position: 3
---

# Verifying changes

Read this when you need to show that a change works before asking for review.

A passing build or an HTTP 200 alone is not proof. Show the user action, the
visible result, any persisted or downstream effect, and the empty, failure or
permission state the change could break. For a bug, reproduce it before you
change anything.

## Choose the journey

The verify-tadoku skill at `.agents/skills/verify-tadoku/SKILL.md` is the
procedure for agents and humans. Start from its feature map,
`.agents/skills/verify-tadoku/references/features/README.md`, find the user
goal your change affects, and read only that area. The map lists entry points,
synthetic accounts, seeded fixtures and known traps.

## Run it in the development environment

Start a DevCLI branch overlay as described in
[Development environment](./environment.md), then drive the journey in a
real browser through the printed `dev url` links. Keep the real login,
frontend, gateway and API path: no mocked API, injected identity or auth bypass.

- Before a write, check that the API response's `X-Dev-Backend` header names
  your overlay rather than the base.
- Kratos, Keto and Flipt are shared. Use disposable identities you own and do not
  change shared accounts, roles or flag policy.
- Finish with `dev down` for your owner and clear the selected links.

Changes to DevCLI routing or synchronization must also pass the gates in
`.dev/acceptance.md`, including affected-service discovery, manifest rendering,
HMR, rebuild latency and owner-scoped teardown.

## Checks before a pull request

Run the checks for what you changed. [Toolchain and repository](./toolchain.md)
explains each command.

| Changed | Run |
| --- | --- |
| Go code | `gofmt -w services/`, `bazel run //:gazelle`, then `bazel build //services/... && bazel test //services/...` |
| Go imports or packages | `bazel run //:gazelle -- -mode=diff`, `./scripts/check-tadoku-api-visibility.sh`, `./tools/ci/check_tadoku_api_provider_deps.sh`, `bazel build //services/tadoku-api/...` |
| SQL queries | `./scripts/generate-sqlc.sh` and commit the full output |
| OpenAPI contract | `./scripts/generate-openapi.sh` and `pnpm api:generate` in `docs/`; commit both outputs |
| Frontend | In `frontend/`: `pnpm --filter <app> exec tsc --noEmit`, `pnpm --filter <app> lint`, then `pnpm build` |
| Docs | In `docs/`: `pnpm build` |

## Evidence on the pull request

`.agents/skills/verify-tadoku/references/pr-evidence.md` describes how to
publish evidence. In short:

- Record the tested commit and any uncommitted diff, owner and branch, selected
  URLs, the backend identities you observed, the steps, and expected versus
  actual results. Name skipped tests, TLS bypasses and anything unverified.
- Attach screenshots, and a short recording for multi-step workflows, directly
  to the pull request with `gh pr comment --attach`. Inspect the media first and
  confirm the posted attachments render.
- Keep media and one-off capture scripts out of commits. Never include
  passwords, cookies, tokens or Secret contents.
- If the code changes after recording, rerun the affected checks or label the
  evidence with the commit it proves.

---
slug: /
sidebar_position: 1
title: Start here
description: What Tadoku is, how the monorepo is laid out, and which page to read for each development task.
---

# Start here

Read this when you are new to the repository or need to find the right page for a task.

Tadoku is a foreign-language immersion tracker and contest platform. Learners
log reading and listening, join contests with registration windows and compare
scores on leaderboards. This site documents how the software is built, run in
the development environment and verified.

## Repository map

| Path | Contains |
| --- | --- |
| `services/` | Go services built with Bazel: `tadoku-api` (the only backend), supporting services and shared packages in `services/common/` |
| `frontend/` | pnpm workspace: webv2, auth and admin, the styleguides and Paper playground, and the `ui` and `paper-ui` design systems |
| `docs/` | This Docusaurus site; pages live in `docs/docs/` |
| `k8s/` | Development-only GitOps base for Argo CD in `k8s/dev/base/` |
| `infra/` | Kratos identity schema and Keto namespace fixtures for backend tests (`infra/dev/ory/`) and a development kubeconfig helper |
| `jobs/` | Container jobs |
| `scripts/` | Code generation, the Tadoku API visibility check, and development seed and secret bootstrap scripts |
| `tools/` | CI check scripts and the pinned oapi-codegen module |
| `.dev/` | DevCLI configuration, deployable definitions and the development runbook |
| `.agents/` | Repository agent skills, including `verify-tadoku` |

## Where to go

| Task | Page |
| --- | --- |
| Install tools, build, test and run code generators | [Toolchain and repository](./develop/toolchain.md) |
| Run your branch against the shared development environment | [Development environment](./develop/environment.md) |
| Prove a change works and publish evidence | [Verifying changes](./develop/verifying-changes.md) |
| Commit, open a pull request, ship a migration or report a bug | [Contributing workflow](./develop/contributing.md) |
| Understand the running system and request paths | [System overview](./architecture/index.md) |
| Work with sign-in, roles and bans | [Authentication and authorization](./architecture/authorization.md) |
| Call one service from another | [Service-to-service auth](./architecture/service-tokens.md) |
| Gate behavior behind a flag | [Feature flags](./architecture/feature-flags.md) |
| Read past design decisions | [Decisions](./adr.md) |
| Change backend behavior | [Tadoku API](./tadoku-api/index.md) |
| Change a frontend application | [Frontend overview](./frontend/index.md) |
| Build a Paper screen | [Paper composition](./frontend/paper-composition.md) |
| Operate the development base or repair data | [Operations](./operations/index.md) |
| Look up a public endpoint | [API reference](./api/index.md) |

## For agents

- `AGENTS.md` at the repository root is the entry point. It holds the rules
  for every change and names the page that holds each area's rules.
- Every page has a one-sentence `description` in its frontmatter and opens with
  a "Read this when …" line, so you can choose pages without reading them fully.
- Repository paths are written as inline code relative to the repository root,
  for example `services/tadoku-api/spec/openapi.yaml`.

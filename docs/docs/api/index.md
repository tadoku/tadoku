---
title: API Reference
description: Public HTTP API contracts for Tadoku services
slug: /api/
---

# API Reference

Tadoku's public HTTP APIs are documented from the single
[Tadoku API contract](https://github.com/tadoku/tadoku/blob/main/services/tadoku-api/spec/openapi.yaml).
The four sections below are filtered public views of that source. The reference
reflects `main`; native and proxied operations retain the same public URLs.

| Domain | Version | Production base URL | Reference |
| --- | --- | --- | --- |
| Immersion | 2.0.0 | `https://tadoku.app/api/immersion/` | [Browse endpoints](./immersion/immersion-api) |
| Content | 1.0.0 | `https://tadoku.app/api/content/` | [Browse endpoints](./content/content-api) |
| Profile | 1.0.0 | `https://tadoku.app/api/profile/` | [Browse endpoints](./profile/profile-api) |
| Authorization | 1.0.0 | `https://tadoku.app/api/authz/` | [Browse endpoints](./authorization/authz-api) |

## Authentication

Some read operations are available anonymously. Operations marked with
`cookieAuth` require a valid `ory_kratos_session` cookie and may also require an
administrator role. The response definitions on each endpoint describe the
expected authorization failures.

This documentation displays request schemas and generated code examples, but
does not send requests from the browser. Interactive requests from the GitHub
Pages origin need a separately reviewed cross-origin authentication policy.

## Scope

Only public, externally routed API specifications are published here. Internal
service-to-service contracts are intentionally excluded from this site.

---
title: API Reference
description: Public HTTP API contracts for Tadoku services, their URLs, and which credentials the gateway accepts, including for third-party clients.
slug: /api/
---

# API Reference

Read this when you call the Tadoku API or build a client for it.

Tadoku's public HTTP APIs are documented from the single
[Tadoku API contract](https://github.com/tadoku/tadoku/blob/main/services/tadoku-api/spec/openapi.yaml).
Tadoku API serves every operation. The sections below are filtered public
views of that one contract, grouped by domain, and reflect `main`.

| Domain | Version | Production base URL | Reference |
| --- | --- | --- | --- |
| Immersion | 2.0.0 | `https://tadoku.app/api/internal/immersion/` | [Browse endpoints](./immersion/immersion-api) |
| Content | 1.0.0 | `https://tadoku.app/api/internal/content/` | [Browse endpoints](./content/content-api) |
| Profile | 1.0.0 | `https://tadoku.app/api/internal/profile/` | [Browse endpoints](./profile/profile-api) |
| Authorization | 1.0.0 | `https://tadoku.app/api/internal/authz/` | [Browse endpoints](./authorization/authz-api) |
| Identity | 1.0.0 | `https://tadoku.app/api/internal` | [Browse endpoints](./identity/identity-api) |
| Email | 1.0.0 | `https://tadoku.app/api/internal` | [Browse endpoints](./email/email-api) |

## URLs

Append an operation's path to its domain's base URL. For example, `POST /logs`
in the Immersion reference is `POST https://tadoku.app/api/internal/immersion/logs`.
These are the routes the Tadoku website calls, and `/api/internal` is the only
API prefix. A path such as `/api/immersion/logs` is not an API route: the
website answers it with its HTML `404` page.

Identity registration and login return `501` for a tenant that selects the
legacy v1 account app; see [Identity accounts and sessions](../tadoku-api/identity).
The email routes also accept tenant access tokens.

## Authentication

The gateway in front of Tadoku API authenticates a request in one of two ways:

| Request | Result |
| --- | --- |
| No credentials | Runs as an anonymous guest. Operations without `cookieAuth` work. Operations that need a signed-in user return `401`. |
| `tadoku_session` cookie | Runs as that user. The identity API sets the cookie when the user signs in on the Tadoku website. Some operations also require the administrator role. |

Apart from tenant access tokens on the email routes, the gateway reads no other
credential. It ignores a session token sent as `X-Session-Token`, so the request
runs as a guest. On other routes, it rejects a request that has an
`Authorization` header and no session cookie with `401`. The response
definitions on each endpoint describe the expected authorization failures.

This documentation displays request schemas and generated code examples, but
does not send requests from the browser. Interactive requests from the GitHub
Pages origin need a separately reviewed cross-origin authentication policy.

### Third-party clients

Tadoku has no supported authentication for third-party clients, such as
e-reader plugins and scripts. There are no API keys, personal access tokens or
OAuth clients, and no session tokens for native apps. Operations that need a
signed-in user, such as submitting a log, accept only the session cookie from
signing in on the Tadoku website.

Anonymous read operations need no credentials:

```sh
curl https://tadoku.app/api/internal/immersion/contests/latest-official
```

## Scope

Only public, externally routed API specifications are published here. Internal
service-to-service contracts are intentionally excluded from this site.

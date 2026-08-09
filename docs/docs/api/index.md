---
title: API Reference
description: Public HTTP API contracts for Tadoku services
slug: /api/
---

# API Reference

Tadoku's public HTTP APIs are documented directly from the OpenAPI contracts
used by the services. The reference reflects the current contracts on the
`main` branch.

| Service | Version | Production base URL | Reference | Source |
| --- | --- | --- | --- | --- |
| Immersion | 2.0.0 | `https://tadoku.app/api/immersion/` | [Browse endpoints](./immersion/immersion-api) | [OpenAPI YAML](https://github.com/tadoku/tadoku/blob/main/services/immersion-api/http/rest/openapi/api.yaml) |
| Content | 1.0.0 | `https://tadoku.app/api/content/` | [Browse endpoints](./content/content-api) | [OpenAPI YAML](https://github.com/tadoku/tadoku/blob/main/services/content-api/http/rest/openapi/api.yaml) |
| Profile | 1.0.0 | `https://tadoku.app/api/profile/` | [Browse endpoints](./profile/profile-api) | [OpenAPI YAML](https://github.com/tadoku/tadoku/blob/main/services/profile-api/http/rest/openapi/api.yaml) |
| Authorization | 1.0.0 | `https://tadoku.app/api/authz/` | [Browse endpoints](./authorization/authz-api) | [OpenAPI YAML](https://github.com/tadoku/tadoku/blob/main/services/authz-api/http/rest/openapi/api.yaml) |

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


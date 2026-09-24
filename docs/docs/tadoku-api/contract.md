---
title: Contract and OpenAPI
description: The canonical Tadoku API OpenAPI contract, the documentation views built from it, its compatibility rules, request decoding and code generation.
sidebar_position: 3
---

# Contract and OpenAPI

Read this when you change `services/tadoku-api/spec/openapi.yaml`, a request or
response shape, or generated HTTP code.

## Canonical contract

- `services/tadoku-api/spec/openapi.yaml` is the one canonical contract for
  public and callback operations.
- Each operation records its documentation source (`x-tadoku-source`) and its
  exposure (`x-tadoku-exposure`, `public` or `callback`). Callback metadata
  records the route's direct caller without making the route public.
- Operation IDs and component names carry their source prefix to avoid
  collisions.
- Equivalent path templates share one canonical path. Each operation keeps its
  documented path and operation ID in `x-legacy-path` and
  `x-legacy-operation-id`, so the documentation views keep their original
  parameter names. Wire URLs do not change.

## Documentation views

- `x-tadoku-sources` defines the documentation views.
  `docs/scripts/api-contract.mjs` builds them from the canonical contract at
  documentation build time. They keep the four sections and URLs of the
  [API reference](../api/index.md); they are views, not four independent
  contracts.
- Manually registered health and metrics endpoints are inventoried in the
  contract's `x-tadoku-operational-surfaces` metadata, not rendered as product
  API.
- Contract tests (`services/tadoku-api/spec/openapi_test.go` and
  `docs/scripts/api-contract.test.mjs`) verify that the contract is valid, that
  server generation covers its operations and the published view counts.

## Compatibility

Every operation must preserve its documented inputs, outputs and business
behavior, including response status, headers, field shapes, nullability, empty
results, filtering, ordering and limits. Prove it with HTTP golden cases
against the production router, using real authentication, authorization and
persistence; see [HTTP golden cases](./http-e2e.md#http-golden-cases).

## Request decoding

- Request bodies use the generated JSON decoder.
- Handlers pass an absent optional body to the application as zero values
  instead of returning a literal 400 response. The operation authorizes first,
  and feature validation turns missing required fields into a 400 response.
- Do not add endpoint-specific middleware, replace request bodies or otherwise
  bypass the generated decoder, for example to alter empty-body behavior.
- Do not add XML or form adapters to work around generated request decoding.

## Code generation

After changing the contract, run from the repository root:

```sh
./scripts/generate-openapi.sh
```

- It runs oapi-codegen through Bazel and writes checked-in shared DTOs and
  standard-library strict-server bindings to
  `services/tadoku-api/generated/openapi/`. CI fails if the output is stale.
- The isolated oapi-codegen module under `tools/oapi-codegen` generates every
  canonical component schema. The server codegen configs
  (`spec/server-codegen.yaml` and `spec/callback-server-codegen.yaml`) limit
  registration to operations this application owns and separate routes with
  different HTTP authentication boundaries.
- Generated routes register through boundary-specific registrars: business
  routes use JWT authentication and the ban gate, callback routes use callback
  credentials. Both keep the shared deadlines and observability.
- The Bazel-built binary has no Go module build-info header, so read the
  oapi-codegen version from `tools/oapi-codegen/go.mod`, not from the generated
  file header.

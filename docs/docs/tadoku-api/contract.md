---
title: Contract and OpenAPI
description: The canonical Tadoku API OpenAPI contract, the documentation views built from it, its compatibility rules, request decoding and code generation.
sidebar_position: 3
---

# Contract and OpenAPI

Read this when you change the OpenAPI contract in
`services/tadoku-api/spec/openapi/`, a request or response shape, or generated
HTTP code.

## Canonical contract

- `services/tadoku-api/spec/openapi.yaml` is the one canonical contract for
  public and callback operations. Code generation, the contract tests and the
  documentation build read it.
- It is generated from the fragments in `services/tadoku-api/spec/openapi/`.
  Edit the fragments, never `openapi.yaml`; see
  [Contract fragments](#contract-fragments).
- Each operation records its documentation source (`x-tadoku-source`) and its
  exposure (`x-tadoku-exposure`, `public` or `callback`). Callback metadata
  records the route's direct caller without making the route public.
- Operation IDs and component names carry their source prefix to avoid
  collisions.
- Equivalent path templates share one canonical path. Each operation keeps its
  documented path and operation ID in `x-legacy-path` and
  `x-legacy-operation-id`, so the documentation views keep their original
  parameter names. Wire URLs do not change.

## Contract fragments

Each fragment holds one part of the contract and stays at or under about 16 KB:

| Fragment in `services/tadoku-api/spec/openapi/` | Holds | Starts with |
| --- | --- | --- |
| `root.yaml` | `openapi`, `info`, `servers`, `x-tadoku-operational-surfaces` and `x-tadoku-sources` | No section header |
| `paths/<source>-<area>.yaml` | The path items of one API area, such as `services/tadoku-api/spec/openapi/paths/immersion-logs.yaml` | `paths:` |
| `schemas/<source>[-<area>].yaml` | The component schemas of one source or area | `components:` with `schemas:` under it |
| `components.yaml` | Security schemes and shared responses | `components:` |

- A fragment is a partial OpenAPI document. `$ref` values keep the canonical
  `#/components/...` form and resolve in the generated `openapi.yaml`.
- `./scripts/generate-openapi.sh` writes `openapi.yaml` by concatenating the
  fragments in the order of its `PATH_FRAGMENTS` and `SCHEMA_FRAGMENTS` lists,
  without each fragment's section header lines. That order is the contract
  order and sets the operation order of the documentation views.
- Add a new fragment to those lists in contract order. The script fails when a
  fragment is not listed or does not start with its section header.
- When a fragment grows past about 16 KB, split it along its API areas.

## Documentation views

- `x-tadoku-sources` defines the documentation views.
  `docs/scripts/api-contract.mjs` builds them from the canonical contract at
  documentation build time. They keep the domain sections and URLs of the
  [API reference](../api/index.md); they are views, not four independent
  contracts.
- Manually registered health and metrics endpoints are inventoried in the
  contract's `x-tadoku-operational-surfaces` metadata, not rendered as product
  API.
- A source's `exposure` decides whether its view is published. Every
  operation's `x-tadoku-exposure` must equal its source's `exposure`.
- Contract tests (`services/tadoku-api/spec/openapi_test.go` and
  `docs/scripts/api-contract.test.mjs`) verify that the contract is valid, that
  server generation covers its operations, that each view holds exactly its
  source's operations and that operation and source exposures agree.
- After the documentation build, `docs/scripts/api-publication.test.mjs`
  checks that the generated and built reference has one page per operation in
  the public views and one introduction per public view. Apart from tag and
  schema pages in a view's directory, it allows no other page.

## Compatibility

Every operation must preserve its documented inputs, outputs and business
behavior, including response status, headers, field shapes, nullability, empty
results, filtering, ordering and limits. Prove it with HTTP golden cases
against the production router, using real authentication, authorization and
persistence; see [HTTP golden cases](./http-e2e.md#http-golden-cases).

Profile and contest-profile language scores are ordered by score descending,
then language code ascending when scores are equal.

## Request decoding

- Request bodies use the generated JSON decoder.
- A shared limit caps every request body at 64 KiB before the decoder reads it,
  so an oversized body cannot consume decoder memory ahead of the operation's
  actor authorization. A body over the limit returns 413 with an empty body.
  Like the 401 from authentication, 413 is a boundary response the contract does
  not document per operation.
- CMS page and post routes allow 1 MiB because they carry a whole hand-authored
  HTML or Markdown document on every save. Every other payload is bounded by
  `maxLength` and `maxItems`; the largest is a platform scoring rule set of about
  8 KB.
- Handlers pass an absent optional body to the application as zero values
  instead of returning a literal 400 response. The operation authorizes first,
  and feature validation turns missing required fields into a 400 response.
- Do not add endpoint-specific middleware, replace request bodies or otherwise
  bypass the generated decoder, for example to alter empty-body behavior. The
  shared byte cap is the only wrapper around a request body.
- Do not add XML or form adapters to work around generated request decoding.

## Code generation

After changing a contract fragment, run from the repository root:

```sh
./scripts/generate-openapi.sh
```

- It bundles the fragments into `services/tadoku-api/spec/openapi.yaml`, then
  runs oapi-codegen through Bazel and writes checked-in shared DTOs and
  standard-library strict-server bindings to
  `services/tadoku-api/generated/openapi/`. CI fails if the bundle or the
  generated code is stale.
- Then run `pnpm api:generate` in `docs/` to regenerate the API reference, and
  `pnpm --filter api-client generate` in `frontend/` to regenerate the
  [frontend API client](../frontend/api-client.md). CI fails if either is stale.
- The isolated oapi-codegen module under `tools/oapi-codegen` generates every
  canonical component schema. The server codegen configs
  (`spec/server-codegen.yaml` and `spec/callback-server-codegen.yaml`) limit
  registration to operations this application owns and separate routes with
  different HTTP authentication boundaries.
- Generated routes register through boundary-specific registrars: business
  routes use JWT authentication and the ban gate, callback routes use callback
  credentials. Both keep the shared deadlines and observability. Email GET routes also select
  a separate token registrar for a verified, tenant-bound `tdk_at_` bearer.
  Invalid access tokens fail before request decoding. All other credentials
  on those routes use the ordinary JWT and ban pipeline.
- A format describes every value a field carries. When a response field's
  values do not all satisfy one, such as an account's stored email, drop the
  format so the generated Go field is a plain type that cannot reject a stored
  value while the response is written. Map the value to its wire form in the
  handler instead.
- The Bazel-built binary has no Go module build-info header, so read the
  oapi-codegen version from `tools/oapi-codegen/go.mod`, not from the generated
  file header.

---
title: Tadoku API client
description: The frontend client generated from the Tadoku API contract, what its response validators check, how applications call it, its exceptions, and how to regenerate and verify it.
---

# Tadoku API client

Read this when frontend code calls Tadoku API, or after you change the API contract.

`frontend/packages/api-client` is generated from
`services/tadoku-api/spec/openapi.yaml` by the pinned `@hey-api/openapi-ts`
version in its `package.json`. It provides:

- request and response types, in `src/generated/types.gen.ts`;
- one function per operation, in `src/generated/sdk.gen.ts`, for example
  `identitySessionGet()` or `immersionLogCreate()`;
- Zod validators, in `src/generated/zod.gen.ts`. Each operation validates its
  success response before the caller sees it;
- each operation's documented success statuses and whether each has a body, in
  `src/generated/responses.gen.ts`, written by
  `frontend/packages/api-client/scripts/generate-responses.mjs`.

Every frontend calls Tadoku API through this package. Do not write a request
URL, request or response type, or response schema for a Tadoku API operation by
hand.

## Regenerate and verify

After changing the contract and running `./scripts/generate-openapi.sh`, run from
`frontend/`:

```sh
pnpm --filter api-client generate          # rewrites packages/api-client/src/generated/
pnpm --filter api-client check:generated   # regenerates, then fails on any change from the commit
pnpm --filter api-client typecheck
pnpm --filter api-client test       # replays recorded API responses through the client
```

Generation needs Node 22.18 or later. Commit the complete output and never edit
`src/generated/` by hand. The `verify-frontend-api-client.yaml` workflow runs
`check:generated` when the contract, the recorded HTTP golden responses or the
package change, then typechecks and tests the client.

`frontend/packages/api-client/test/contract.test.ts` sends every recorded
response in `services/tadoku-api/e2e/testdata/` through the generated operations.
Each documented success must validate, and must fail under an undocumented
status or with its body added or removed. Each non-empty documented error body
must match its error validator. It also sends malformed answers and values that
break the contract's constraints, which must fail.

## What a response must satisfy

- A success uses a status the operation documents, with a body exactly when that
  status documents one. Anything else, such as `201` for a `200` operation or an
  empty `200` where the contract has JSON, fails. Error statuses reach the caller
  with their status, with or without a body.
- Validators check JSON types, required properties, nullability, enum values,
  formats, lengths, patterns, ranges, item counts and objects the contract
  closes with `additionalProperties: false`.
- Three rules follow the API rather than Zod's defaults: lengths count
  characters (code points), not UTF-16 units, so a 25-emoji display name fits 32;
  `format: email` accepts what Go's `net/mail` writes, such as `reader@localhost`;
  date-times accept any RFC 3339 offset. `openapi-ts.config.mjs` holds them.
- Every answer is parsed as JSON. An HTML page fails.
- Validation does not transform data: a call returns the JSON as sent. Use only
  the properties the types declare.
- Operations do not validate error bodies. Code that reads one parses it with the
  operation's error validator, for example `zIdentityErrorResponse`.
- Requests are not validated in the browser; the API validates them.

The contract is the only source of these rules. When the API's real answers
differ from it, fix the contract in `services/tadoku-api/spec/openapi/`; never
patch it in the generator configuration. Many documented JSON errors arrive
without a body, and callers fall back to the status. The two Oathkeeper callback
operations are excluded, because no frontend calls them.

## Call the client

Pass `baseUrl` on every call: the client has no default server. Pass `headers` as
a plain object: operations with a JSON body spread it, which drops a `Headers`
instance. Thin adapters own base URLs, cookie and branch forwarding, `Origin`,
credentials, caching and timeouts.

| Application | Adapters |
| --- | --- |
| authv2 | `frontend/apps/authv2/src/lib/session.server.ts` reads the session. `sendIdentity()` in `frontend/apps/authv2/src/lib/identity.ts` sends every account form, from the browser or the native form route, and reads error codes with `zIdentityErrorResponse`. |
| webv3 | `frontend/apps/webv3/src/lib/home.server.ts` makes anonymous home-page reads; `404` means no contest. |
| webv2, admin, auth | `frontend/packages/ui/identity.ts` reads the session and logs out. `browserApi()` and `apiData()` in `frontend/packages/ui/api.ts` call operations from React Query hooks and throw `Error(<status>)` for API errors. |

Some contract properties are optional only because request bodies or drafts share
the schema: contest and registration IDs, page HTML, post publish times and
announcement IDs. Where a screen needs one, `requireFields()` in
`frontend/packages/ui/api.ts`, or `identified()` in webv3's home data, fails like
a malformed answer when it is missing.

## Exceptions

- The log export, `GET /immersion/logs/export`, is a download link that the
  browser navigates to.
- The auth application's Kratos flows use Kratos's own client and contract.
- Form validation and error messages belong to each form.
- Manual Playwright journeys read API responses directly to drive and check
  flows.

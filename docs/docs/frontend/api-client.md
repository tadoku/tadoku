---
title: Tadoku API client
description: The frontend client generated from the Tadoku API contract, what its response validators check, how applications call it, the contract gaps it patches, its exceptions, and how to regenerate and verify it.
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
  success response before the caller sees it.

Every frontend calls Tadoku API through this package. Do not write a request
URL, request or response type, or response schema for a Tadoku API operation by
hand.

## Regenerate and verify

After changing the contract and running `./scripts/generate-openapi.sh`, run from
`frontend/`:

```sh
pnpm --filter api-client generate   # rewrites packages/api-client/src/generated/
pnpm --filter api-client typecheck
pnpm --filter api-client test       # replays recorded API responses through the client
```

Generation needs Node 22.18 or later. Commit the complete output and never edit
`src/generated/` by hand. The `verify-frontend-api-client.yaml` workflow
regenerates the client when the contract, the recorded HTTP golden responses or
the package change, fails on any difference, then typechecks and tests it.

`frontend/packages/api-client/test/contract.test.ts` sends every recorded
response in `services/tadoku-api/e2e/testdata/` through the generated operations.
Each documented success must validate and each non-empty documented error body
must match its error validator. It also sends malformed answers that must fail.

## What a response must satisfy

- Validators check JSON types, required properties, nullability, enum values,
  integers and objects the contract closes with `additionalProperties: false`.
- They do not check lengths, formats, patterns, ranges or item counts. Those
  describe what the API accepts; Zod counts UTF-16 units rather than characters,
  and stored data predates some limits.
- Every answer is parsed as JSON. An empty success body is validated as absent,
  so an HTML page, an empty body or `204` fails unless the operation documents a
  success without a body.
- Validation does not transform data: a call returns the JSON as sent. Use only
  the properties the types declare.
- Operations do not validate error bodies. Code that reads one parses it with the
  operation's error validator, for example `zIdentityErrorResponse`.
- Requests are not validated in the browser; the API validates them.

## Contract gaps patched for the frontend

`frontend/packages/api-client/openapi-ts.config.mjs` patches the contract where
the API's actual behavior differs from it:

| Patch | Reason |
| --- | --- |
| Contest `language_code_allow_list` and contest view `allowed_languages` are nullable | The API answers `null` for a contest without a language allow list. |
| The feature-access `environment` includes `test` | Test tenants evaluate flags in the Flipt test environment. |
| `POST /identity/logout` takes a JSON object body | The API rejects identity writes without `Content-Type: application/json`. |

Remove a patch once the contract documents the behavior. The recorded responses
also show post lists of up to 100 items, above the documented 50, and many
documented JSON errors answered without a body; the validators do not check item
counts, and callers fall back to the status. The two Oathkeeper callback
operations are excluded, because no frontend calls them.

## Call the client

Pass `baseUrl` on every call: the client has no default server. Thin adapters own
base URLs, cookie and branch forwarding, `Origin`, credentials, caching and
timeouts.

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

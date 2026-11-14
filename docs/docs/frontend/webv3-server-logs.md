---
title: Webv3 server logs
description: The structured JSON log lines Webv3's server writes for upstream failures, rendering errors and errors browsers report, their fields, which statuses count as failures, what they can never contain, the request deadline and the signed-out read cache.
---

# Webv3 server logs

Read this when you record a failure on Webv3's server or query its logs.

`frontend/apps/webv3/src/lib/log.server.ts` records upstream failures, server
rendering errors and errors browsers report, plus one startup marker. Each failure writes one JSON line to
stdout at level `ERROR`. Once per boot, after the startup checks in
`frontend/apps/webv3/server/plugins/runtime-env.ts` pass, the server writes one
`INFO` line with `event` `startup` and `msg` `server started`, and no other
fields, so every rollout leaves a line. All lines use the field names of Tadoku API's `log/slog` JSON handler
(`services/tadoku-api/cmd/tadoku-api/main.go`,
`services/tadoku-api/transport/http/observability.go`,
`services/tadoku-api/infra/observability/scoring.go`), so one Loki query reads
both.

| Field | Value |
| --- | --- |
| `time`, `level`, `msg` | As slog writes them; `msg` is a fixed string per event |
| `event` | `upstream_failure`, `ssr_error`, `client_error` or `startup` |
| `correlation_id` | The request's ID |
| `host` | The API gateway host this request reads from |
| `operation` | The generated API operation that failed |
| `status` | The HTTP status, when there is one |
| `latency` | The upstream call's duration in integer nanoseconds, as slog encodes a `time.Duration` |
| `error_type` | `http_status`, `invalid_response`, `timeout` or `network` for an upstream failure; the error's class for a thrown server error, or `other` for a class outside the sink's list; the report's kind for a browser error |
| `issue_paths` | The first 20 Zod issue paths of an invalid response, as dotted paths; `$` is the response root |
| `error`, `stack`, `path`, `build_id`, `user_agent` | A browser error's message, stack, page path, build and browser; see [Browser errors](#browser-errors) |

`recordUpstreamFailure` and `recordSsrError` accept only their typed shapes and copy
named fields, so their events cannot carry cookies, tokens, request bodies, URLs, error
messages, Zod messages or received values. `operation` is a compile-time literal.
A Zod path segment is kept only when it is a field name declared by an object
schema the API client exports, so a record key and a field of an inline nested
object become `*`.

`withRequestLog` in `frontend/apps/webv3/src/start.ts` gives every Start request an
ID through `AsyncLocalStorage`, which records read without callers passing it. A
well-formed ingress `X-Request-ID` of 1 to 128 letters, digits, `_` or `-` is
kept; anything else is replaced with a generated ID. Every upstream call sends
the request's ID to Tadoku API as `X-Request-Id`, which the API logs as its
`correlation_id` (`services/tadoku-api/transport/http/observability.go`), so
both services' lines for one request share it. Records made outside a Start
request, such as from a Nitro route in `frontend/apps/webv3/server/routes/`, have
no `correlation_id` or `host`.

## Upstream failures

Every API call the server makes takes its options from `sessionRequest` or
`publicReadRequest` in `frontend/apps/webv3/src/lib/session.server.ts`, which name
the operation. `log.server.ts` registers one error interceptor on the generated
client, which sees each failed call once, so adapters record nothing themselves.

| Outcome | Recorded as |
| --- | --- |
| Deadline or per-call timeout, including during the body | `timeout` |
| Connection failure, or a body cut off while reading | `network` |
| Any error status other than those below, such as 429 or any 5xx | `http_status` |
| A success that fails Zod validation, is not JSON, or has a status or body the contract doesn't document | `invalid_response` |

Statuses 400, 401, 403, 404 and 409 are not recorded: adapters turn them into an
answer the page explains, such as invalid input, a rejected session that
`readAsVisitor` retries signed out, a refusal, a missing item or an account being
deleted. A documented success such as 204 never reaches the interceptor.

The same middleware records an SSR error for a response of status 500 or above
and for a rejected request, unless the client has aborted it. An error React
recovers from while rendering, such as one thrown by a page component, keeps the
response's status, usually 200, and is recorded as an SSR error with its class in
`error_type`. A request records at most one SSR error, so a render error and the
500 it causes are one line. A call to a server function this build doesn't have
answers 409 and records nothing; see
[status codes](webv3-status-codes.md). The framework also writes its own unstructured
errors to stderr, thrown messages included; that stream is not covered by this
contract.

## Browser errors

The browser reports an error that only happens on the page, which no server
record can see, and the server writes each report as one `client_error` line.
`frontend/apps/webv3/src/lib/client-errors.ts` reports these kinds, named in
`error_type`:

| Kind | Reported when |
| --- | --- |
| `error` | An exception is uncaught (`window` `error`) |
| `unhandledrejection` | A promise rejection is unhandled |
| `hydration` | React recovers from an error while hydrating, such as a mismatch with the server's HTML (`onRecoverableError` in `frontend/apps/webv3/src/client.tsx`) |
| `chunk` | Vite cannot load a code chunk (`vite:preloadError`), as after a release removed it |
| `route` | An error reaches the route error view, `ErrorView` in `frontend/apps/webv3/src/components/StatusViews.tsx` |

A page sends each message once and at most five reports. A page whose main read
failed is not reported, nor what React reports while hydrating around it: the
server has already recorded it as an upstream failure and an SSR error. The
report carries the message, the stack, the page's path without its query, and
`build_id`, the commit the image was built from (`VITE_BUILD_ID`, `dev` outside a
published image).

The browser sends the report with `navigator.sendBeacon` to `POST /client-error`,
`frontend/apps/webv3/server/routes/client-error.post.ts`. It is a Nitro route
outside the Start request middleware, so its line has no `correlation_id` or
`host`, and anyone can call it. It reads at most 16 KiB of body and validates the
report with Zod. An oversized body, invalid JSON or an invalid report answers 400
and writes nothing, and no input can make it answer 5xx. Each server process
writes at most 120 lines a minute and answers 429 beyond that. The server, not
the browser, redacts the message, stack and path: it drops the query and fragment
of every absolute or relative URL, with any URL nested in them, keeping a stack
frame's `:line:column`, and replaces each email address with `[email]`. The line
copies only the named fields and never reads cookies. The message and stack are
browser text, which can still quote other page content.

## Querying

Production and its production branches run on the `lke-prod` cluster, whose pod
logs reach the homelab Loki with each JSON line in the record's `body`.
Development-cluster logs, dev-cli branches included, do not reach it. Webv3's
records for the canonical deployment:

```logql
{k8s_cluster_name="lke-prod", service_namespace="tdk-prod-frontend-webv3", service_name="frontend-webv3"}
  |~ "upstream_failure|ssr_error"
  | json body="body" | line_format "{{.body}}" | json
```

Browser errors, grouped by kind and message:

```logql
sum by (error_type, error) (count_over_time(
  {k8s_cluster_name="lke-prod", service_namespace="tdk-prod-frontend-webv3", service_name="frontend-webv3"}
    |= "client_error"
    | json body="body" | line_format "{{.body}}" | json | event="client_error" [1h]))
```

Tadoku API's lines use `service_namespace="tdk-prod-tadoku-api"` and
`service_name="tadoku-api"` with the same parsing.

## Request deadline

Every GET request has an overall deadline for its upstream calls, set in
`frontend/apps/webv3/src/lib/session.server.ts`. Upstream calls still running at
the deadline are aborted, each is recorded as one upstream failure with
`error_type` `timeout`, and the page renders with those sections marked as
failed. A request whose calls have finished keeps rendering.

## Signed-out read cache

`publicReadRequest` answers a public read from memory for ten seconds when the
incoming request carries no `tadoku_session` or `tadoku_session_<tenant>` cookie.
The latest official contest, standings pages, CMS pages and blog reads use it. A
request with any such cookie, even an empty one, always reads the API. A
signed-out page can start from a read up to ten seconds old, which the browser
then keeps for its query's own stale time.

An entry is keyed by the absolute API URL, whose host selects the tenant, so a
tenant only reads its own entries. Only a 200 answer that passed response
validation is stored, at most 500 entries with the oldest evicted first. The cache
lives in one process, so replicas don't share it and a restart empties it.

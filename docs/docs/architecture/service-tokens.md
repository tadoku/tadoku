---
sidebar_position: 4
title: Service-to-service authentication
description: How Tadoku API exchanges its Kubernetes service account token for short-lived Oathkeeper service JWTs to call internal services such as Flipt.
---

# Service-to-Service Authentication

Read this when you add an authenticated call from Tadoku API to an internal service, change an Oathkeeper token-exchange rule, or debug a failed token exchange.

## Overview

Tadoku API calls Flipt for feature-flag evaluation and feature-access management.
It never sends its Kubernetes credential to Flipt. Instead it exchanges that
credential at Oathkeeper for a short-lived service JWT scoped to one target, and
Oathkeeper guards the target routes with that JWT.

```text
tadoku-api ── projected SA token ──▶ Oathkeeper /token-exchange/<target>/<caller>
           ◀──── service JWT ─────── (returned by token-reflector)
tadoku-api ── service JWT ─────────▶ Oathkeeper /flipt… ──▶ Flipt
```

## Development topology

| Component | Namespace | Role |
| --- | --- | --- |
| Tadoku API, service account `tadoku-api` | `tdk-dev-tadoku-api` | Caller; mounts a projected service account token. |
| Oathkeeper, `oathkeeper-proxy:4455` and `oathkeeper-api:4456` | `tdk-dev-oathkeeper` | Verifies the caller, mints service JWTs, guards the Flipt routes. |
| token-reflector (`services/token-reflector/`) | `tdk-dev-token-reflector` | Serves the cluster JWKS, checks the caller's service account, returns the minted JWT. |
| Flipt | `tdk-dev-flipt` | Target. Its network policy admits only Oathkeeper pods and the `monitoring` namespace. |

Manifests are in `k8s/dev/base/services/`, `k8s/dev/base/oathkeeper/` and
`k8s/dev/base/flipt/`; exchange and target rules are in
`k8s/dev/base/oathkeeper/rules.yaml`.

## Projected service account token

The Tadoku API deployment disables the automounted token and mounts a projected
one at `/var/run/secrets/tokens/token`, with audience `tadoku-api` and a 3600 s
expiry. The audience identifies the caller, not the target. Tadoku API reads the
path from `API_SERVICE_ACCOUNT_TOKEN_PATH` (default
`/var/run/secrets/tokens/token`) and the exchange base URL from
`API_OATHKEEPER_URL` (development: `http://oathkeeper-proxy.tdk-dev-oathkeeper:4455`).

## Token exchange

The client lives in `services/common/client/s2s/`.

1. The client reads the projected token and sends
   `GET <API_OATHKEEPER_URL>/token-exchange/<target>/<caller>` with
   `Authorization: Bearer <projected token>`.
2. Oathkeeper's `jwt` authenticator verifies the token against
   `http://token-reflector.tdk-dev-token-reflector/jwks`, which proxies the
   Kubernetes API server's `/openid/v1/jwks`, and requires audience `tadoku-api`.
3. The `remote_json` authorizer sends the verified `sub` and the rule's
   `expected_subject` (`system:serviceaccount:tdk-dev-tadoku-api:tadoku-api`) to
   token-reflector's `/authorize-service-account`, which returns `403` unless they
   are equal.
4. The `id_token` mutator mints a service JWT with Oathkeeper's signing key and
   forwards the request to token-reflector.
5. token-reflector returns the minted token as
   `{"access_token": "<jwt>", "token_type": "Bearer", "expires_in": <seconds>}`,
   where `expires_in` counts down to the token's `exp`.
6. The client rejects an empty token, a non-bearer type or a non-positive
   expiry. It caches each target's token until five minutes before expiry.

`s2s.NewAuthTransport(client, "<target>/<caller>", base)` wraps an
`http.RoundTripper` so every request gets a current bearer token; the exchange
follows the request's cancellation and deadline. `services/tadoku-api/cmd/tadoku-api/main.go`
builds separate Flipt HTTP clients for `flipt-evaluation/tadoku-api` and
`flipt-management/tadoku-api`.

## Service JWT claims

| Claim | Value |
| --- | --- |
| `sub` | Caller service account, `system:serviceaccount:<namespace>:<name>` |
| `aud` | Array with the target: `flipt-evaluation` or `flipt-management` |
| `type` | `service` |
| `iss`, `iat`, `exp` | Set by Oathkeeper; the issuer is `http://oathkeeper-api/` |

## Target routes

Oathkeeper forwards `API_FLIPT_URL` (`…:4455/flipt`) and
`API_FLIPT_MANAGEMENT_URL` (`…:4455/flipt-management`) requests to Flipt only for
the exact paths and methods listed in `rules.yaml`. Each such rule uses a `jwt`
authenticator against Oathkeeper's own JWKS
(`http://oathkeeper-api.tdk-dev-oathkeeper:4456/.well-known/jwks.json`), requires
the matching `target_audience`, and strips the prefix before forwarding to
`flipt.tdk-dev-flipt:8080`.

Tadoku API itself rejects service JWTs with `401`; they are valid only on
Oathkeeper-guarded internal routes.

To add a target, follow the same pattern: an `s2s:` exchange rule with the
caller's audience, the exact `expected_subject` and an `id_token` mutator setting
`aud` to the target; target rules that require that audience; and a client
wrapped with `s2s.NewAuthTransport`.

## Failure modes

| Failure | Result |
| --- | --- |
| Projected token missing or unreadable | The client fails before sending the exchange. |
| Projected token invalid, expired or with the wrong audience | Oathkeeper returns `401`. |
| Valid token from the wrong service account | token-reflector returns `403`; Oathkeeper returns `403`. |
| token-reflector cannot fetch the Kubernetes JWKS | `/jwks` returns `502`, so Oathkeeper cannot verify caller tokens. |
| Exchange returns a non-`200` status or an invalid body | The client returns an error and the call to the target is not sent. |
| Service JWT missing or with the wrong audience on a Flipt route | Oathkeeper returns `401`. |

In Tadoku API, a failed exchange leaves feature-flag evaluation on its safe
defaults without failing startup, and feature-access management operations
return `503`.

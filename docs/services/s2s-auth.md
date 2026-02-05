# Service-to-Service Authentication

This document describes how internal services authenticate with each other using short-lived JWTs issued by Oathkeeper. The same middleware supports both user and service identities.

## Overview

- Services use Kubernetes service account (SA) tokens to request a short-lived JWT from Oathkeeper.
- Oathkeeper validates the SA token via the JWT authenticator, mints a service JWT, and forwards it via the token-reflector.
- Services include the service JWT in `Authorization: Bearer <token>` when calling other internal services.
- The receiving service validates the JWT and attaches a `ServiceIdentity` to the request context.

## Identity Model

The middleware attaches a `domain.Identity` to the request context:

- `UserIdentity` for human users authenticated via Kratos.
- `ServiceIdentity` for internal services authenticated via Kubernetes SA.

The `subject` (JWT `sub`, short for "subject") is the primary identifier:

- User: stable unique user ID from the identity provider.
- Service: `system:serviceaccount:<namespace>:<name>`.

## Token Exchange Flow

1. Service reads its projected SA token at `/var/run/secrets/tokens/token`.
2. Service calls the Oathkeeper token exchange endpoint:
   - `GET http://oathkeeper-proxy.default:4455/token-exchange/<target-service>`
   - `Authorization: Bearer <sa-token>`
3. Oathkeeper validates the SA token using the JWT authenticator and the Kubernetes JWKS served by token-reflector.
4. Oathkeeper mints a service JWT and forwards the request to token-reflector.
5. Token-reflector returns a JSON response:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

6. Caller uses the JWT to call the target service.

## Token Claims

Service JWTs include:

- `sub`: full service account name
- `aud`: array with the target service name
- `type`: `service`
- `namespace`: inferred from `sub`

User JWTs include:

- `sub`: user ID
- `type`: `user`
- `session`: Kratos session traits

## Middleware Responsibilities

Authn:

- `VerifyJWT` verifies the JWT signature (JWKS).
- `Identity` builds `UserIdentity` or `ServiceIdentity` and attaches it to context.

Authz:

- `RequireServiceAudience` enforces the service audience using the configured service name.
  - Each service sets a default via envconfig: `service_name` defaults to the service's name (for example, `immersion-api`).
- `RejectBannedUsers` blocks banned users (except `/current-user/role`).

## Local Development Setup

- Each service runs in its own namespace prefixed with `tdk-`.
- Each service has its own ServiceAccount.
- `token-reflector` runs in `tdk-token-reflector`.
- Oathkeeper validates SA tokens using a JWKS URL served by token-reflector:
  - `http://token-reflector.tdk-token-reflector/jwks`

## Example: immersion-api -> profile-api

1. `immersion-api` requests a token for `profile-api` via Oathkeeper.
2. `immersion-api` calls `http://profile-api.tdk-profile-api/internal/v1/ping` with the JWT.
3. `profile-api` validates the JWT and checks the service audience.

## Quickstart (Calling Another Service)

1. Ensure your config includes `oathkeeper_url` and `service_name` (defaults are set per service).
2. Initialize the S2S client.
3. Use the generated internal client with a custom HTTP transport.

### Using a Generated Internal Client (Preferred)

```go
import (
	"net/http"

	"github.com/tadoku/tadoku/services/common/client/s2s"
	profileclient "github.com/tadoku/tadoku/services/profile-api/http/rest/openapi/internalapi"
)

s2sClient := s2s.NewClient(cfg.OathkeeperURL)
httpClient := &http.Client{
	Transport: s2s.NewAuthTransport(s2sClient, "profile-api", nil),
}

client, err := profileclient.NewClient(
	"http://profile-api.tdk-profile-api",
	profileclient.WithHTTPClient(httpClient),
)
if err != nil {
	return err
}

resp, err := client.InternalPing(ctx)
if err != nil {
	return err
}
_ = resp
```

## Failure Modes

- Invalid audience: `403 Forbidden`
- Missing/invalid JWT: `401 Unauthorized`
- Missing service name configuration: service tokens rejected (`403`)
- Token-reflector unavailable: token exchange returns `502 Bad Gateway`

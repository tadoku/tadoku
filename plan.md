# Health Checks Implementation Plan

## Overview

Replace the basic `GET /ping` "pong" endpoint with two proper health check endpoints following Kubernetes probe best practices:

- **`GET /livez`** — Liveness probe. Shallow check: "is the process alive and not deadlocked?" Returns 200 if the server can respond. No dependency checks. If this fails, Kubernetes restarts the pod.
- **`GET /readyz`** — Readiness probe. Dependency-aware check: "can this instance serve traffic?" Checks critical dependencies (e.g. Postgres). If this fails, the pod is removed from the load balancer but stays alive.

Starting with **content-api**, with shared interfaces and reusable checkers in **common**.

---

## Step 1: Create shared `health` package in `common`

**New file: `services/common/health/health.go`**

Contains the `HealthChecker` interface, response types, **and** reusable checkers (like Postgres) since all services share the same dependencies.

```go
package health

import (
    "context"
    "database/sql"
)

// HealthChecker represents a dependency that can be health-checked.
type HealthChecker interface {
    Name() string
    Check(ctx context.Context) error
}

// CheckResult holds the outcome of a single dependency check.
type CheckResult struct {
    Name   string `json:"name"`
    Status string `json:"status"` // "up" or "down"
    Error  string `json:"error,omitempty"`
}

// ReadyzResponse is the JSON response for the readiness endpoint.
type ReadyzResponse struct {
    Status string        `json:"status"` // "ready" or "not_ready"
    Checks []CheckResult `json:"checks"`
}

// postgresChecker verifies database connectivity using sql.DB.PingContext.
type postgresChecker struct {
    name string
    db   *sql.DB
}

// NewPostgresChecker creates a HealthChecker for a Postgres connection.
// The name parameter identifies this check in the response (e.g. "postgres",
// "postgres-content", "postgres-immersion") so services with multiple
// databases can distinguish them.
func NewPostgresChecker(name string, db *sql.DB) HealthChecker {
    return &postgresChecker{name: name, db: db}
}

func (c *postgresChecker) Name() string                        { return c.name }
func (c *postgresChecker) Check(ctx context.Context) error     { return c.db.PingContext(ctx) }
```

The Postgres checker is in `common` because every service has Postgres and the implementation is identical — only the name and `*sql.DB` differ.

---

## Step 2: Shared HTTP handlers in `common/health`

**New file: `services/common/health/handler.go`**

```go
package health

import (
    "context"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
)

const checkTimeout = 2 * time.Second

// LivezHandler returns a simple 200 OK — the process is alive.
func LivezHandler(c echo.Context) error {
    return c.String(http.StatusOK, "ok")
}

// ReadyzHandler checks all registered HealthCheckers and returns
// 200 if all pass, 503 if any fail.
func ReadyzHandler(checkers []HealthChecker) echo.HandlerFunc {
    return func(c echo.Context) error {
        ctx, cancel := context.WithTimeout(c.Request().Context(), checkTimeout)
        defer cancel()

        response := ReadyzResponse{Status: "ready"}
        allHealthy := true

        for _, checker := range checkers {
            result := CheckResult{Name: checker.Name(), Status: "up"}
            if err := checker.Check(ctx); err != nil {
                result.Status = "down"
                result.Error = err.Error()
                allHealthy = false
            }
            response.Checks = append(response.Checks, result)
        }

        if !allHealthy {
            response.Status = "not_ready"
            return c.JSON(http.StatusServiceUnavailable, response)
        }

        return c.JSON(http.StatusOK, response)
    }
}
```

Key design decisions:
- **2-second timeout** on all dependency checks combined to prevent hanging
- **Iterates all checkers** even if one fails, so the response shows the full picture
- Returns **503** when not ready (standard for readiness probes)

---

## Step 3: Auth model — Optional admin auth middleware

**Requirement:** K8s probes (no JWT) are allowed through. External requests (with JWT, via Oathkeeper gateway) must be admin.

**Approach:** Split the middleware stacks using Echo route groups.

- Health routes (`/livez`, `/readyz`) are registered on the **root** `e` with a lightweight `OptionalAdminAuth` middleware
- Business routes (OpenAPI handlers including `/ping`) are registered on an **`api` group** with the full auth middleware stack

**New file: `services/common/middleware/optional_auth.go`**

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/MicahParks/keyfunc"
    jwtv4 "github.com/golang-jwt/jwt/v4"
    "github.com/labstack/echo/v4"
    "github.com/tadoku/tadoku/services/common/authz/roles"
)

// OptionalAdminAuth creates middleware that allows unauthenticated requests
// (e.g. Kubernetes probes) but requires admin authorization when a JWT
// bearer token is present (e.g. external requests through the API gateway).
func OptionalAdminAuth(jwksURL string, rolesSvc roles.Service) echo.MiddlewareFunc {
    jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
    if err != nil {
        panic(fmt.Errorf("optional auth: unable to fetch jwks: %w", err))
    }

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            auth := c.Request().Header.Get("Authorization")
            if !strings.HasPrefix(auth, "Bearer ") {
                // No JWT — allow through (K8s probe)
                return next(c)
            }

            tokenString := strings.TrimPrefix(auth, "Bearer ")
            token, err := jwtv4.ParseWithClaims(tokenString, &UnifiedClaims{}, jwks.Keyfunc)
            if err != nil || !token.Valid {
                return c.NoContent(http.StatusUnauthorized)
            }

            claims := token.Claims.(*UnifiedClaims)

            // Service tokens with valid JWT are allowed (internal service-to-service)
            if claims.Type == "service" {
                return next(c)
            }

            // User tokens must be admin
            subject := claims.Subject
            if subject == "" || subject == "guest" {
                return c.NoContent(http.StatusUnauthorized)
            }

            roleClaims, err := rolesSvc.ClaimsForSubject(c.Request().Context(), subject)
            if err != nil {
                return c.NoContent(http.StatusServiceUnavailable)
            }
            if roleClaims.Banned || !roleClaims.Admin {
                return c.NoContent(http.StatusForbidden)
            }

            return next(c)
        }
    }
}
```

This reuses the existing `UnifiedClaims` type from `session.go` and the `roles.Service` interface already used by `RolesFromKeto`.

---

## Step 4: Wire up in content-api `main.go`

The key structural change: move auth middleware from `e.Use()` to an `api` group, keeping health routes on the root with different middleware.

```go
e := echo.New()
e.Use(middleware.Recover())

// --- Health endpoints (optional auth — allow K8s probes, require admin if JWT) ---
pgChecker := commonhealth.NewPostgresChecker("postgres", psql)
checkers := []commonhealth.HealthChecker{pgChecker}

optAuth := tadokumiddleware.OptionalAdminAuth(cfg.JWKS, rolesSvc)
e.GET("/livez", commonhealth.LivezHandler, optAuth)
e.GET("/readyz", commonhealth.ReadyzHandler(checkers), optAuth)

// --- Business endpoints (full auth stack) ---
api := e.Group("")
api.Use(tadokumiddleware.Logger([]string{"/ping"}))
api.Use(tadokumiddleware.VerifyJWT(cfg.JWKS))
api.Use(tadokumiddleware.Identity())
api.Use(tadokumiddleware.RolesFromKeto(rolesSvc))
api.Use(tadokumiddleware.RequireServiceAudience(cfg.ServiceName))
api.Use(tadokumiddleware.RejectBannedUsers())
// Sentry middleware also on api group if configured

openapi.RegisterHandlersWithBaseURL(api, server, "")
```

The `EchoRouter` interface that `RegisterHandlersWithBaseURL` accepts is satisfied by both `*echo.Echo` and `*echo.Group`, so this works without touching the generated code.

---

## Step 5: Update Kubernetes deployment

Update `services/content-api/deployments/api.yaml`:

```yaml
readinessProbe:
  httpGet:
    path: /readyz    # was /ping
    port: 8000
  initialDelaySeconds: 10
  periodSeconds: 3
livenessProbe:
  httpGet:
    path: /livez     # was /ping
    port: 8000
  initialDelaySeconds: 10
  periodSeconds: 3
```

---

## Step 6: Keep `/ping` endpoint (backward compatibility)

Keep `/ping` working as-is via the OpenAPI-generated router. It stays on the `api` group with full auth middleware — no changes to `health.go` or the OpenAPI spec.

---

## Step 7: Write tests

**Test the shared handler + checker logic: `services/common/health/handler_test.go`**

- Test `LivezHandler` returns 200 "ok"
- Test `ReadyzHandler` with all healthy checkers returns 200 + `{"status":"ready",...}`
- Test `ReadyzHandler` with a failing checker returns 503 + `{"status":"not_ready",...}`
- Test that the 2s timeout is enforced (mock a slow checker)
- Test `NewPostgresChecker` name is configurable

**Test the optional auth middleware: `services/common/middleware/optional_auth_test.go`**

- Test request with no Authorization header → allowed
- Test request with valid admin JWT → allowed
- Test request with valid non-admin JWT → 403
- Test request with invalid JWT → 401

---

## Step 8: Regenerate BUILD.bazel files

```sh
bazel run //:gazelle
```

---

## What's NOT in scope (for now)

- **immersion-api / profile-api health checks** — will follow the same pattern later, adding their own checkers (Valkey, Kratos, etc.)
- **Startup probes** — can be added later if services develop slow startup
- **Detailed metrics** (latency_ms, uptime, git SHA) — can be added to response types later
- **Removing `/ping`** — kept for backward compatibility
- **No changes to `VerifyJWT` skipper** — health endpoints are on a separate route group, so the skipper is irrelevant for them

---

## File summary

| File | Action |
|------|--------|
| `services/common/health/health.go` | **New** — shared interface, response types, reusable PostgresChecker |
| `services/common/health/handler.go` | **New** — shared LivezHandler + ReadyzHandler |
| `services/common/health/handler_test.go` | **New** — handler + checker tests |
| `services/common/middleware/optional_auth.go` | **New** — OptionalAdminAuth middleware |
| `services/common/middleware/optional_auth_test.go` | **New** — optional auth tests |
| `services/content-api/main.go` | **Edit** — restructure to use route groups, wire up health endpoints |
| `services/content-api/deployments/api.yaml` | **Edit** — point K8s probes at `/livez` and `/readyz` |

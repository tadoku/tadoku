package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/common/domain"
)

const (
	bearerPrefix              = "Bearer "
	echoJWTExtractorLimit     = 20
	echoJWTMissingOrMalformed = "missing or malformed jwt"
	echoJWTInvalidOrExpired   = "invalid or expired jwt"
)

func VerifyJWT(jwksURL string) echo.MiddlewareFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			panic(fmt.Errorf("unable to refresh jwks: %w", err))
		},
	})

	if err != nil {
		panic(fmt.Errorf("unable to fetch jwks: %w", err))
	}

	// Echo v4.10+ moved JWT helpers out of echo/v4/middleware. Keep the v4.9
	// status mapping (400 missing, 401 invalid) so legacy services stay compatible.
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/ping" {
				return next(c)
			}

			tokens, err := bearerTokens(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, echoJWTMissingOrMalformed)
			}

			var lastTokenErr error
			for _, raw := range tokens {
				token, err := jwt.ParseWithClaims(raw, &UnifiedClaims{}, jwks.Keyfunc)
				if err != nil || token == nil || !token.Valid {
					lastTokenErr = err
					continue
				}

				c.Set("user", token)

				return next(c)
			}

			return &echo.HTTPError{
				Code:     http.StatusUnauthorized,
				Message:  echoJWTInvalidOrExpired,
				Internal: lastTokenErr,
			}
		}
	}
}

func bearerTokens(c echo.Context) ([]string, error) {
	values := c.Request().Header.Values(echo.HeaderAuthorization)
	if len(values) == 0 {
		return nil, echo.ErrBadRequest
	}

	prefixLen := len(bearerPrefix)
	result := make([]string, 0)
	for i, value := range values {
		if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], bearerPrefix) {
			result = append(result, value[prefixLen:])
			if i >= echoJWTExtractorLimit-1 {
				break
			}
		}
	}

	if len(result) == 0 {
		return nil, echo.ErrBadRequest
	}

	return result, nil
}

// UnifiedClaims handles both user and service tokens.
type UnifiedClaims struct {
	jwt.RegisteredClaims
	Type      string `json:"type,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Session   struct {
		Identity struct {
			Traits struct {
				DisplayName string `json:"display_name"`
				Email       string
			}
		}
	} `json:"session,omitempty"`
}

func Identity() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var identity domain.Identity = &domain.UserIdentity{
				Subject: "guest",
			}

			if ctx.Get("user") == nil {
				setIdentityContext(ctx, identity)
				return next(ctx)
			}

			token := ctx.Get("user").(*jwt.Token)
			if claims, ok := token.Claims.(*UnifiedClaims); ok && token.Valid {
				switch claims.Type {
				case "service":
					identity = handleServiceToken(claims)
				default:
					identity = handleUserToken(claims)
				}
			}

			setIdentityContext(ctx, identity)

			return next(ctx)
		}
	}
}

func handleServiceToken(claims *UnifiedClaims) domain.Identity {
	name := claims.Subject
	namespace := claims.Namespace
	if parts := strings.Split(claims.Subject, ":"); len(parts) == 4 {
		namespace = parts[2]
		name = parts[3]
	}

	return &domain.ServiceIdentity{
		Subject:   claims.Subject,
		Name:      name,
		Namespace: namespace,
		Audience:  []string(claims.Audience),
	}
}

func handleUserToken(claims *UnifiedClaims) *domain.UserIdentity {
	user := &domain.UserIdentity{
		Email:       claims.Session.Identity.Traits.Email,
		DisplayName: claims.Session.Identity.Traits.DisplayName,
		Subject:     claims.Subject,
		CreatedAt:   claims.IssuedAt.Time,
	}

	return user
}

func setIdentityContext(ctx echo.Context, identity domain.Identity) {
	ctx.SetRequest(ctx.Request().WithContext(
		context.WithValue(ctx.Request().Context(), domain.CtxIdentityKey, identity)))
}

func RejectBannedUsers() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			claims := roles.FromContext(ctx.Request().Context())
			if claims.Authenticated && claims.Err != nil {
				// Fail-open: allow requests to proceed when authorization evaluation is unavailable.
				// Admin-only endpoints will still be blocked by roles.RequireAdmin (ErrAuthzUnavailable).
				ctx.Logger().Errorf(
					"RejectBannedUsers: authorization unavailable (fail-open): subject=%s err=%v",
					claims.Subject,
					claims.Err,
				)
				return next(ctx)
			}
			if claims.Banned {
				return ctx.NoContent(http.StatusForbidden)
			}
			return next(ctx)
		}
	}
}

func RequireServiceAudience(serviceName string) echo.MiddlewareFunc {
	if serviceName == "" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				if domain.ParseServiceIdentity(ctx.Request().Context()) != nil {
					return ctx.NoContent(http.StatusForbidden)
				}
				return next(ctx)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if service := domain.ParseServiceIdentity(ctx.Request().Context()); service != nil {
				for _, aud := range service.Audience {
					if aud == serviceName {
						return next(ctx)
					}
				}
				return ctx.NoContent(http.StatusForbidden)
			}
			return next(ctx)
		}
	}
}

func RequireServiceIdentity() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if domain.ParseServiceIdentity(ctx.Request().Context()) == nil {
				return ctx.NoContent(http.StatusUnauthorized)
			}
			return next(ctx)
		}
	}
}

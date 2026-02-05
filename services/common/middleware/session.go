package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tadoku/tadoku/services/common/domain"
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

	return middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(context echo.Context) bool {
			return context.Path() == "/ping"
		},
		Claims: &UnifiedClaims{},
		KeyFunc: func(token *jwt.Token) (interface{}, error) {
			t, _, err := new(jwtv4.Parser).ParseUnverified(token.Raw, &UnifiedClaims{})
			if err != nil {
				return nil, err
			}
			return jwks.Keyfunc(t)
		},
	})
}

// UnifiedClaims handles both user and service tokens.
type UnifiedClaims struct {
	jwtv4.RegisteredClaims
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

// RoleRepository provides role lookup by email (for config file)
type RoleRepository interface {
	GetRole(email string) string
}

// DatabaseRoleRepository provides role lookup by user ID (for database)
type DatabaseRoleRepository interface {
	GetUserRole(ctx context.Context, userID string) (string, error)
}

func Identity(configRepo RoleRepository, dbRepo DatabaseRoleRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			var identity domain.Identity = &domain.UserIdentity{
				Subject: "guest",
				Role:    domain.RoleGuest,
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
					identity = handleUserToken(ctx, claims, configRepo, dbRepo)
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

func handleUserToken(ctx echo.Context, claims *UnifiedClaims, configRepo RoleRepository, dbRepo DatabaseRoleRepository) *domain.UserIdentity {
	user := &domain.UserIdentity{
		Email:       claims.Session.Identity.Traits.Email,
		DisplayName: claims.Session.Identity.Traits.DisplayName,
		Subject:     claims.Subject,
		CreatedAt:   claims.IssuedAt.Time,
	}

	// First check config file (for dev/admin overrides).
	if configRepo != nil {
		role := configRepo.GetRole(user.Email)
		if role != "user" {
			user.Role = domain.Role(role)
		}
	}

	// Then check database if no special role from config.
	if user.Role == "" || user.Role == domain.RoleUser || user.Role == domain.RoleGuest {
		if dbRepo != nil {
			dbRole, err := dbRepo.GetUserRole(ctx.Request().Context(), user.Subject)
			if err == nil && dbRole != "user" {
				user.Role = domain.Role(dbRole)
			}
		}
	}

	// Default to user if no special role found.
	if user.Role == "" || user.Role == domain.RoleGuest {
		user.Role = domain.RoleUser
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
			if user := domain.ParseUserIdentity(ctx.Request().Context()); user != nil {
				if user.Role == domain.RoleBanned && ctx.Path() != "/current-user/role" {
					return ctx.NoContent(http.StatusForbidden)
				}
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

package serviceauth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
)

// contextKey is used for storing service auth info in context
type contextKey string

const callingServiceKey contextKey = "calling_service"

// GetCallingService returns the calling service name from the context
// Returns empty string if the request is not from a service (i.e., user request)
func GetCallingService(c echo.Context) string {
	if caller, ok := c.Request().Context().Value(callingServiceKey).(string); ok {
		return caller
	}
	return ""
}

// ServiceAuth creates middleware that only accepts valid service JWTs
// Use this for internal-only endpoints that should never be called by users
func ServiceAuth(validator *TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			callingService, err := extractAndValidateServiceToken(c, validator)
			if err != nil {
				log.Printf("service auth failed: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
			}

			// Store calling service in context
			ctx := context.WithValue(c.Request().Context(), callingServiceKey, callingService)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// ServiceOrUserAuth creates middleware that accepts either service JWT or user JWT
// For service JWTs: validates the token and sets calling service in context
// For user JWTs: falls through to session middleware behavior
func ServiceOrUserAuth(validator *TokenValidator, roleRepo middleware.RoleRepository, dbRepo middleware.DatabaseRoleRepository) echo.MiddlewareFunc {
	sessionMiddleware := middleware.Session(roleRepo, dbRepo)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this looks like a service token
			authHeader := c.Request().Header.Get("Authorization")
			if tokenString, found := strings.CutPrefix(authHeader, "Bearer "); found {
				// Try to validate as service token
				callingService, err := validator.Validate(tokenString)
				if err == nil {
					// Valid service token - store in context and proceed
					ctx := context.WithValue(c.Request().Context(), callingServiceKey, callingService)
					c.SetRequest(c.Request().WithContext(ctx))
					return next(c)
				}
				// Not a valid service token - fall through to session middleware
			}

			// Let session middleware handle it (user JWT or no auth)
			return sessionMiddleware(next)(c)
		}
	}
}

// extractAndValidateServiceToken extracts and validates a service token from the request
func extractAndValidateServiceToken(c echo.Context, validator *TokenValidator) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	tokenString, found := strings.CutPrefix(authHeader, "Bearer ")
	if !found {
		return "", &authError{"missing or invalid Authorization header"}
	}

	return validator.Validate(tokenString)
}

// authError represents an authentication error
type authError struct {
	message string
}

func (e *authError) Error() string {
	return e.message
}

// IsServiceRequest returns true if the current request is from a service
func IsServiceRequest(c echo.Context) bool {
	return GetCallingService(c) != ""
}

// IsUserRequest returns true if the current request is from a user
func IsUserRequest(c echo.Context) bool {
	return domain.ParseSession(c.Request().Context()) != nil
}

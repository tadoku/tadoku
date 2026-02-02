package serviceauth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenExpiry is the default expiry duration for service tokens
// Short-lived tokens (30s) minimize replay window without needing a JTI cache
const TokenExpiry = 30 * time.Second

// ClockSkewLeeway is the tolerance for clock drift between services
const ClockSkewLeeway = 10 * time.Second

// ServiceClaims represents the JWT claims for service-to-service authentication
type ServiceClaims struct {
	jwt.RegisteredClaims
}

// Valid validates time-based claims with clock skew tolerance
// This is needed because jwt/v4 doesn't have WithLeeway (that's a v5 feature)
func (c ServiceClaims) Valid() error {
	now := time.Now()

	// Check expiry with leeway
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Add(ClockSkewLeeway)) {
		return jwt.NewValidationError("token is expired", jwt.ValidationErrorExpired)
	}

	// Check not-before with leeway
	if c.NotBefore != nil && now.Before(c.NotBefore.Add(-ClockSkewLeeway)) {
		return jwt.NewValidationError("token is not valid yet", jwt.ValidationErrorNotValidYet)
	}

	// Check issued-at with leeway (token shouldn't be from the future)
	if c.IssuedAt != nil && now.Before(c.IssuedAt.Add(-ClockSkewLeeway)) {
		return jwt.NewValidationError("token used before issued", jwt.ValidationErrorIssuedAt)
	}

	return nil
}

// NewServiceClaims creates claims for a service-to-service token
// Per RFC 7523, sub is set to the service name (same as iss)
func NewServiceClaims(issuer, audience string, now time.Time) ServiceClaims {
	return ServiceClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   issuer,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenExpiry)),
		},
	}
}

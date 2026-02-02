package serviceauth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenExpiry is the default expiry duration for service tokens
const TokenExpiry = 5 * time.Minute

// ServiceClaims represents the JWT claims for service-to-service authentication
type ServiceClaims struct {
	jwt.RegisteredClaims
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

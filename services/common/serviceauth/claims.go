package serviceauth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// SubjectService is the subject claim value for service-to-service tokens
const SubjectService = "service"

// TokenExpiry is the default expiry duration for service tokens
const TokenExpiry = 5 * time.Minute

// ServiceClaims represents the JWT claims for service-to-service authentication
type ServiceClaims struct {
	jwt.RegisteredClaims
}

// NewServiceClaims creates claims for a service-to-service token
func NewServiceClaims(issuer, audience string, now time.Time) ServiceClaims {
	return ServiceClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   SubjectService,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenExpiry)),
		},
	}
}

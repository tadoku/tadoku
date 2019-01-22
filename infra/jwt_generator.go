package infra

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tadoku/api/usecases"
)

// NewJWTGenerator intializes a new JWTGenerator
func NewJWTGenerator(signingKey string) usecases.JWTGenerator {
	return &jwtGenerator{signingKey: signingKey}
}

// JWTGenerator makes it easy to generate JWT tokens that expire in a given duration
type jwtGenerator struct {
	signingKey string
}

// NewToken generates a signed JWT token
func (g *jwtGenerator) NewToken(expiresIn time.Duration, claims interface{}) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		struct {
			claims interface{}
			jwt.StandardClaims
		}{
			claims,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(expiresIn).Unix(),
			},
		},
	)

	return token.SignedString([]byte(g.signingKey))
}

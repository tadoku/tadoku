package infra

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/tadoku/api/usecases"
)

// NewJWTGenerator intializes a new JWTGenerator
func NewJWTGenerator(signingKey string, clock usecases.Clock) usecases.JWTGenerator {
	return &jwtGenerator{signingKey: signingKey, clock: clock}
}

// JWTGenerator makes it easy to generate JWT tokens that expire in a given duration
type jwtGenerator struct {
	signingKey string
	clock      usecases.Clock
}

// NewToken generates a signed JWT token
func (g *jwtGenerator) NewToken(lifetime time.Duration, src usecases.SessionClaims) (string, error) {
	claims := jwtClaims{
		SessionClaims: src,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: g.clock.Now().Add(lifetime).Unix(),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(g.signingKey))
}

type jwtClaims struct {
	usecases.SessionClaims
	jwt.StandardClaims
}

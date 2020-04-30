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
func (g *jwtGenerator) NewToken(lifetime time.Duration, src usecases.SessionClaims) (string, int64, error) {
	expiresAt := g.clock.Now().Add(lifetime).Unix()

	claims := jwtClaims{
		SessionClaims: src,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	result, err := token.SignedString([]byte(g.signingKey))

	return result, expiresAt, err
}

type jwtClaims struct {
	usecases.SessionClaims
	jwt.StandardClaims
}

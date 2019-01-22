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
func (g *jwtGenerator) NewToken(lifetime time.Duration, data map[string]interface{}) (string, error) {

	claims := make(jwt.MapClaims, len(data)+1)
	for k, v := range data {
		claims[k] = v
	}

	claims["exp"] = time.Now().Add(lifetime).Unix()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(g.signingKey))
}

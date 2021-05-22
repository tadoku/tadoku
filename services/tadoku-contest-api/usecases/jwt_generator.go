//go:generate gex mockgen -source=jwt_generator.go -package usecases -destination=jwt_generator_mock.go

package usecases

import (
	"time"
)

// JWTGenerator makes it easy to generate JWT tokens that expire in a given duration
type JWTGenerator interface {
	NewToken(lifetime time.Duration, claims SessionClaims) (token string, expiresAt int64, err error)
}

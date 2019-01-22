//go:generate gex mockgen -source=jwt.go -package usecases -destination=jwt_mock.go

package usecases

import (
	"time"
)

// JWTGenerator makes it easy to generate JWT tokens that expire in a given duration
type JWTGenerator interface {
	NewToken(expiresIn time.Duration, claims ...interface{}) (token string, err error)
}

package app

import (
	"time"

	"github.com/tadoku/tadoku/services/identity-api/infra"
	"github.com/tadoku/tadoku/services/identity-api/usecases"
)

// Interactors is a collection of all repositories
type Interactors struct {
	Session usecases.SessionInteractor
	User    usecases.UserInteractor
}

// NewInteractors initializes all repositories
func NewInteractors(
	r *Repositories,
	jwtGenerator usecases.JWTGenerator,
	sessionLength time.Duration,
) *Interactors {
	passwordHasher := infra.NewPasswordHasher()

	return &Interactors{
		Session: usecases.NewSessionInteractor(
			r.User,
			passwordHasher,
			jwtGenerator,
			sessionLength,
		),
		User: usecases.NewUserInteractor(r.User, passwordHasher),
	}
}

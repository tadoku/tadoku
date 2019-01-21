package app

import (
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/usecases"
)

// Interactors is a collection of all repositories
type Interactors struct {
	Session usecases.SessionInteractor
}

// NewInteractors initializes all repositories
func NewInteractors(r *Repositories) *Interactors {
	return &Interactors{
		Session: usecases.NewSessionInteractor(r.User, infra.NewPasswordHasher()),
	}
}

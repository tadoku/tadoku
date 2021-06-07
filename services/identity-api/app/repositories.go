package app

import (
	"github.com/tadoku/tadoku/services/identity-api/interfaces/rdb"
	r "github.com/tadoku/tadoku/services/identity-api/interfaces/repositories"
	"github.com/tadoku/tadoku/services/identity-api/usecases"
)

// Repositories is a collection of all repositories
type Repositories struct {
	User usecases.UserRepository
}

// NewRepositories initializes all repositories
func NewRepositories(sh rdb.SQLHandler) *Repositories {
	return &Repositories{
		User: r.NewUserRepository(sh),
	}
}

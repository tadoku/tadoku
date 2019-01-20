package interactors

import (
	"github.com/tadoku/api/domain"
	r "github.com/tadoku/api/usecases/repositories"
)

// SessionInteractor contains all business logic for sessions
type SessionInteractor interface {
	CreateUser(user domain.User) error
	// CreateSession(email, password, deviceID string) (user domain.User, token string, err error)
}

// NewSessionInteractor instantiates SessionInteractor with all dependencies
func NewSessionInteractor(userRepository r.UserRepository) SessionInteractor {
	return &sessionInteractor{
		UserRepository: userRepository,
	}
}

type sessionInteractor struct {
	UserRepository r.UserRepository
}

func (si *sessionInteractor) CreateUser(user domain.User) error {
	return nil
}

package interactors

import (
	"github.com/srvc/fail"
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
		userRepository: userRepository,
	}
}

type sessionInteractor struct {
	userRepository r.UserRepository
}

func (si *sessionInteractor) CreateUser(user domain.User) error {
	if user.ID != 0 {
		return fail.Errorf("User with an ID (%v) could not be created.", user.ID)
	}

	return si.userRepository.Store(user)
}

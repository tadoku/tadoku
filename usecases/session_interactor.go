package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// SessionInteractor contains all business logic for sessions
type SessionInteractor interface {
	CreateUser(user domain.User) error
	CreateSession(email, password string) (user domain.User, token string, err error)
}

// NewSessionInteractor instantiates SessionInteractor with all dependencies
func NewSessionInteractor(userRepository UserRepository) SessionInteractor {
	return &sessionInteractor{
		userRepository: userRepository,
	}
}

type sessionInteractor struct {
	userRepository UserRepository
}

func (si *sessionInteractor) CreateUser(user domain.User) error {
	if user.ID != 0 {
		return fail.Errorf("User with an ID (%v) could not be created.", user.ID)
	}

	err := si.userRepository.Store(user)
	return fail.Wrap(err)
}

func (si *sessionInteractor) CreateSession(email, password string) (domain.User, string, error) {
	user, err := si.userRepository.FindByEmail(email, true)
	if err != nil {
		return domain.User{}, "", fail.Wrap(err)
	}

	token := ""

	return user, token, nil
}

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
func NewSessionInteractor(userRepository UserRepository, passwordHasher PasswordHasher) SessionInteractor {
	return &sessionInteractor{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}

type sessionInteractor struct {
	userRepository UserRepository
	passwordHasher PasswordHasher
}

func (si *sessionInteractor) CreateUser(user domain.User) error {
	if user.ID != 0 {
		return fail.Errorf("User with an ID (%v) could not be created.", user.ID)
	}

	if user.NeedsHashing() {
		var err error
		user.Password, err = si.passwordHasher.Hash(user.Password)
		if err != nil {
			return fail.Wrap(err)
		}
	}

	err := si.userRepository.Store(user)
	return fail.Wrap(err)
}

func (si *sessionInteractor) CreateSession(email, password string) (domain.User, string, error) {
	user, err := si.userRepository.FindByEmail(email, true)
	if err != nil {
		return domain.User{}, "", fail.Wrap(err)
	}

	if !si.passwordHasher.Compare(user.Password, password) {
		return domain.User{}, "", fail.Wrap(fail.New("invalid password supplied"), fail.WithIgnorable())
	}

	token := ""

	return user, token, nil
}

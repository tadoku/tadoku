package usecases

import (
	"time"

	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// SessionInteractor contains all business logic for sessions
type SessionInteractor interface {
	CreateUser(user domain.User) error
	CreateSession(email, password string) (user domain.User, token string, err error)
}

// NewSessionInteractor instantiates SessionInteractor with all dependencies
func NewSessionInteractor(
	userRepository UserRepository,
	passwordHasher PasswordHasher,
	jwtGenerator JWTGenerator,
	sessionLength time.Duration,
) SessionInteractor {
	return &sessionInteractor{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		jwtGenerator:   jwtGenerator,
		sessionLength:  sessionLength,
	}
}

type sessionInteractor struct {
	userRepository UserRepository
	passwordHasher PasswordHasher
	jwtGenerator   JWTGenerator
	sessionLength  time.Duration
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
	user, err := si.userRepository.FindByEmail(email)
	if err != nil {
		return domain.User{}, "", fail.Wrap(err)
	}

	if !si.passwordHasher.Compare(user.Password, password) {
		return domain.User{}, "", fail.Wrap(fail.New("invalid password supplied"), fail.WithIgnorable())
	}

	token, err := si.jwtGenerator.NewToken(si.sessionLength, user)
	if err != nil {
		return domain.User{}, "", fail.Wrap(err)
	}

	return user, token, nil
}

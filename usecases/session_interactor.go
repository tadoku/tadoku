//go:generate gex mockgen -source=session_interactor.go -package usecases -destination=session_interactor_mock.go

package usecases

import (
	"time"

	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// ErrUserDoesNotExist for when a user could not be found
var ErrUserDoesNotExist = fail.New("user does not exist")

// SessionInteractor contains all business logic for sessions
type SessionInteractor interface {
	CreateSession(email, password string) (user domain.User, token string, expiresAt int64, err error)
	RefreshSession(user domain.User) (latestUser domain.User, token string, expiresAt int64, err error)
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

func (si *sessionInteractor) CreateSession(email, password string) (domain.User, string, int64, error) {
	user, err := si.userRepository.FindByEmail(email)
	if err != nil {
		return domain.User{}, "", 0, domain.WrapError(err)
	}

	if user.ID == 0 {
		return domain.User{}, "", 0, domain.WrapError(ErrUserDoesNotExist, fail.WithIgnorable())
	}

	if !si.passwordHasher.Compare(user.Password, password) {
		return domain.User{}, "", 0, domain.WrapError(domain.ErrPasswordIncorrect, fail.WithIgnorable())
	}

	claims := SessionClaims{User: &user}
	token, expiresAt, err := si.jwtGenerator.NewToken(si.sessionLength, claims)
	if err != nil {
		return domain.User{}, "", 0, domain.WrapError(err)
	}

	return user, token, expiresAt, nil
}

func (si *sessionInteractor) RefreshSession(user domain.User) (domain.User, string, int64, error) {
	user, err := si.userRepository.FindByEmail(user.Email)
	if err != nil {
		return domain.User{}, "", 0, domain.WrapError(err)
	}

	if user.ID == 0 {
		return domain.User{}, "", 0, domain.WrapError(ErrUserDoesNotExist, fail.WithIgnorable())
	}

	claims := SessionClaims{User: &user}
	token, expiresAt, err := si.jwtGenerator.NewToken(si.sessionLength, claims)
	if err != nil {
		return domain.User{}, "", 0, domain.WrapError(err)
	}

	return user, token, expiresAt, nil
}

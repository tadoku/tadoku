//go:generate gex mockgen -source=user_interactor.go -package usecases -destination=user_interactor_mock.go

package usecases

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/domain"
)

// UserInteractor contains all business logic for users
type UserInteractor interface {
	CreateUser(user domain.User) error
	UpdatePassword(email string, currentPassword, newPassword string) error
	UpdateProfile(user domain.User) error
}

// NewUserInteractor instantiates UserInteractor with all dependencies
func NewUserInteractor(
	userRepository UserRepository,
	passwordHasher PasswordHasher,
) UserInteractor {
	return &userInteractor{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}

type userInteractor struct {
	userRepository UserRepository
	passwordHasher PasswordHasher
}

func (i *userInteractor) CreateUser(user domain.User) error {
	if user.ID != 0 {
		return fail.Errorf("User with an ID (%v) could not be created.", user.ID)
	}

	if user.NeedsHashing() {
		var err error
		user.Password, err = i.passwordHasher.Hash(user.Password)
		if err != nil {
			return domain.WrapError(err)
		}
	}

	err := i.userRepository.Store(&user)
	return domain.WrapError(err)
}

func (i *userInteractor) UpdatePassword(email string, currentPassword, newPassword string) error {
	user, err := i.userRepository.FindByEmail(email)
	if err != nil {
		return domain.WrapError(err)
	}

	if user.ID == 0 {
		return domain.WrapError(ErrUserDoesNotExist, fail.WithIgnorable())
	}

	if !i.passwordHasher.Compare(user.Password, currentPassword) {
		return domain.WrapError(domain.ErrPasswordIncorrect, fail.WithIgnorable())
	}

	user.Password, err = i.passwordHasher.Hash(domain.Password(newPassword))
	if err != nil {
		return domain.WrapError(err)
	}

	err = i.userRepository.UpdatePassword(&user)

	return domain.WrapError(err)
}

func (i *userInteractor) UpdateProfile(user domain.User) error {
	return i.userRepository.Store(&user)
}

package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// UserService is responsible for user related when they're logged in
type UserService interface {
	Register(ctx Context) error
	UpdatePassword(ctx Context) error
	UpdateProfile(ctx Context) error
}

// NewUserService initializer
func NewUserService(userInteractor usecases.UserInteractor) UserService {
	return &userService{
		UserInteractor: userInteractor,
	}
}

type userService struct {
	UserInteractor usecases.UserInteractor
}

func (u *userService) Register(ctx Context) error {
	user := &domain.User{}
	err := ctx.Bind(user)
	if err != nil {
		return domain.WrapError(err)
	}

	user.Role = domain.RoleUser
	user.Preferences = &domain.Preferences{}

	err = u.UserInteractor.CreateUser(*user)
	if err != nil {
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// UserUpdatePasswordBody is the data that's needed to update your password
type UserUpdatePasswordBody struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (u *userService) UpdatePassword(ctx Context) error {
	user, err := ctx.User()
	if err != nil {
		return domain.WrapError(err)
	}

	b := &UserUpdatePasswordBody{}
	err = ctx.Bind(b)
	if err != nil {
		return domain.WrapError(err)
	}

	err = u.UserInteractor.UpdatePassword(user.Email, b.CurrentPassword, b.NewPassword)
	if err != nil {
		if err == domain.ErrPasswordIncorrect {
			return ctx.NoContent(401)
		}

		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// UserUpdateProfileBody is the data that's needed to update your profile
type UserUpdateProfileBody struct {
	DisplayName string `json:"display_name"`
}

func (u *userService) UpdateProfile(ctx Context) error {
	user, err := ctx.User()
	if err != nil {
		return domain.WrapError(err)
	}

	b := &UserUpdateProfileBody{}
	err = ctx.Bind(b)
	if err != nil {
		return domain.WrapError(err)
	}

	user.DisplayName = b.DisplayName

	err = u.UserInteractor.UpdateProfile(*user)
	if err != nil {
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusOK)
}

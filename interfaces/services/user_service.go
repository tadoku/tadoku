package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// UserService is responsible for user related when they're logged in
type UserService interface {
	UpdatePassword(ctx Context) error
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
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusOK)
}

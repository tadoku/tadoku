package usecases

import (
	"github.com/tadoku/api/domain"
)

// UserRepository handles User related interactions
type UserRepository interface {
	Store(user domain.User) error
	// @TODO: make this a pointer
	FindByID(id uint64) (domain.User, error)
	FindByEmail(email string, withPassword bool) (*domain.User, error)
}

package usecases

import (
	"github.com/tadoku/api/domain"
)

// UserRepository handles User related interactions
type UserRepository interface {
	Store(user domain.User) error
	FindByID(id int) domain.User
}

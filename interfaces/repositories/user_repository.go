package repositories

import (
	"github.com/tadoku/api/domain"
)

// UserRepository handles User related interactions
type UserRepository interface {
	Store(user domain.User)
	FindById(id int) domain.User
}

//go:generate gex mockgen -source=repositories.go -package usecases -destination=repositories_mock.go

package usecases

import (
	"github.com/tadoku/tadoku/services/identity-api/domain"
)

// UserRepository handles User related database interactions
type UserRepository interface {
	Store(user *domain.User) error
	UpdatePassword(user *domain.User) error
	FindByID(id uint64) (domain.User, error)
	FindByEmail(email string) (domain.User, error)
}

//go:generate gex mockgen -source=repositories.go -package usecases -destination=repositories_mock.go

package usecases

import (
	"github.com/tadoku/api/domain"
)

// UserRepository handles User related interactions
type UserRepository interface {
	Store(user domain.User) error
	FindByID(id uint64) (domain.User, error)
	FindByEmail(email string) (domain.User, error)
}

// ContestRepository handles Contest related interactions
type ContestRepository interface {
	Store(contest domain.Contest) error
}

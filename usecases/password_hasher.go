//go:generate gex mockgen -source=password_hasher.go -package usecases -destination=password_hasher_mock.go

package usecases

import (
	"github.com/tadoku/api/domain"
)

// PasswordHasher is for hashing passwords without having to worry about the implementation
type PasswordHasher interface {
	Hash(unhashed domain.Password) (hashed domain.Password, err error)
	Compare(hash domain.Password, original string) bool
}

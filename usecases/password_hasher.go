//go:generate gex mockgen -source=password_hasher.go -package usecases -destination=password_hasher_mock.go

package usecases

// PasswordHasher is for hashing passwords without having to worry about the implementation
type PasswordHasher interface {
	Hash(string) (string, error)
	Compare(first, second string) bool
}

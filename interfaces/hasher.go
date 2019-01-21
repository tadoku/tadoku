//go:generate gex mockgen -source=hasher.go -package interfaces -destination=hasher_mock.go

package interfaces

// Hasher is for hashing strings without having to worry about the implementation
type Hasher interface {
	Hash(string) (string, error)
}

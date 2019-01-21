package interfaces

// Hasher is for hashing strings without having to worry about the implementation
type Hasher interface {
	Hash(string) (string, error)
}

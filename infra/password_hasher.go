package infra

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/interfaces"
	"golang.org/x/crypto/bcrypt"
)

// NewPasswordHasher initializes a new password hasher with sane defaults
func NewPasswordHasher() interfaces.Hasher {
	return &passwordHasher{cost: bcrypt.DefaultCost}
}

type passwordHasher struct {
	cost int
}

func (h *passwordHasher) Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), h.cost)
	if err != nil {
		return "", fail.Wrap(err)
	}

	return string(hash), nil
}

func (h *passwordHasher) Compare(first, second string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(first), []byte(second))
	return err == nil
}

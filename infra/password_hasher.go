package infra

import (
	"github.com/srvc/fail"
	"github.com/tadoku/api/usecases"
	"golang.org/x/crypto/bcrypt"
)

// NewPasswordHasher initializes a new password hasher with sane defaults
func NewPasswordHasher() usecases.PasswordHasher {
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

func (h *passwordHasher) Compare(hash, original string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(original))
	return err == nil
}

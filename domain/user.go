package domain

import (
	"github.com/srvc/fail"
)

// Password is a safe version of a string that will never be exposed when marshalled into JSON
type Password string

// User holds everything related to a user's account data
type User struct {
	ID          uint64       `json:"id" db:"id"`
	Email       string       `json:"email" db:"email"`
	DisplayName string       `json:"display_name" db:"display_name" valid:"utfletternum,required,runelength(2|18)"`
	Role        Role         `json:"role" db:"role"`
	Preferences *Preferences `json:"preferences" db:"preferences"`

	// Password max runelength is an arbitrary high number that in theory should never be hit
	Password         Password `json:"password" db:"password" valid:"runelength(6|99999999)"`
	isPasswordHashed bool
}

// Users is a collection of users
type Users []User

// ErrUserMissingPassword for when a new user is created without a password
var ErrUserMissingPassword = fail.New("a new user must have a password")

// NeedsHashing tells you if the password is in need of being hashed
func (u *User) NeedsHashing() bool {
	return u.Password != "" && !u.isPasswordHashed
}

// IsAdmin returns true when the user has the administration role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// Validate a user
func (u User) Validate() (bool, error) {
	if u.ID == 0 && u.Password == "" {
		return false, ErrUserMissingPassword
	}

	return true, nil
}

// MarshalJSON prevents the password from being exported into something client-facing
func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

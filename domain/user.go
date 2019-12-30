package domain

import (
	"unicode"

	"github.com/srvc/fail"
)

// Password is a safe version of a string that will never be exposed when marshalled into JSON
type Password string

// User holds everything related to a user's account data
type User struct {
	ID          uint64       `json:"id" db:"id"`
	Email       string       `json:"email" db:"email"`
	DisplayName string       `json:"display_name" db:"display_name" valid:"required"`
	Role        Role         `json:"role" db:"role"`
	Preferences *Preferences `json:"preferences" db:"preferences"`

	// Password max runelength is an arbitrary high number that in theory should never be hit
	Password         Password `json:"password" db:"password" valid:"optional,runelength(6|99999999)"`
	isPasswordHashed bool
}

// Users is a collection of users
type Users []User

// ErrUserMissingPassword for when a new user is created without a password
var ErrUserMissingPassword = fail.New("a new user must have a password")

// ErrDisplayNameInvalid for when a display name is incorrect
var ErrDisplayNameInvalid = fail.New("a display name should consist of letters, numbers and -_")

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
	if !validateDisplayName(u.DisplayName) {
		return false, ErrDisplayNameInvalid
	}

	if u.ID == 0 && u.Password == "" {
		return false, ErrUserMissingPassword
	}

	return true, nil
}

// MarshalJSON prevents the password from being exported into something client-facing
func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

func validateDisplayName(name string) bool {
	for _, c := range name {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c) || c == '_' || c == '-' || c == ' ') {
			return false
		}
	}

	length := len(name)
	if !(length >= 2 && length <= 18) {
		return false
	}

	return true
}

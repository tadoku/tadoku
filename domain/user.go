package domain

// Password is a safe version of a string that will never be exposed when marshalled into JSON
type Password string

// User holds everything related to a user's account data
type User struct {
	ID               uint64       `json:"id" db:"id"`
	Email            string       `json:"email" db:"email"`
	DisplayName      string       `json:"display_name" db:"display_name"`
	Password         Password     `json:"password" db:"password"`
	Role             Role         `json:"role" db:"role"`
	Preferences      *Preferences `json:"preferences" db:"preferences"`
	isPasswordHashed bool
}

// Users is a collection of users
type Users []User

// NeedsHashing tells you if the password is in need of being hashed
func (u *User) NeedsHashing() bool {
	return u.Password != "" && !u.isPasswordHashed
}

// MarshalJSON prevents the password from being exported into something client-facing
func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}

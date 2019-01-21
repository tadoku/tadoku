package domain

// User holds everything related to a user's account data
type User struct {
	ID               uint64       `json:"id" db:"id"`
	Email            string       `json:"email" db:"email"`
	DisplayName      string       `json:"display_name" db:"display_name"`
	Password         string       `json:"password" db:"password"`
	Role             Role         `json:"role" db:"role"`
	Preferences      *Preferences `json:"preferences" db:"preferences"`
	isPasswordHashed bool
}

// Users is a collection of users
type Users []User

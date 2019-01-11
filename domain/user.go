package domain

type User struct {
	ID          uint64      `json:"id" db:"id"`
	Email       string      `json:"email" db:"email"`
	DisplayName string      `json:"display_name" db:"display_name"`
	Password    string      `json:"password" db:"password"`
	Role        Role        `json:"role" db:"role"`
	Preferences Preferences `json:"preferences" db:"preferences"`
}

type Users []User

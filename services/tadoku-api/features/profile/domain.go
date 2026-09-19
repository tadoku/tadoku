// Package profile owns administrator-facing user profile operations.
package profile

type CachedUser struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

type User struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
	Role        string
}

type UserList struct {
	Users     []User
	TotalSize int
}

package domain

type SessionToken struct {
	Subject     string
	DisplayName string
	Email       string
	Role        Role
}

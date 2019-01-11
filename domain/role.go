package domain

type Role int

const (
	RoleDisabled Role = iota
	RoleUser
	RoleAdmin
)

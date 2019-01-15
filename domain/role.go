package domain

// Role is an enum with the access level of a user
type Role int

// These are all the possible values for Role
const (
	RoleDisabled Role = iota
	RoleUser
	RoleAdmin
)

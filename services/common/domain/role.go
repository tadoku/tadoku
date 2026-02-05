package domain

import "context"

type Role string

var RoleAdmin = Role("admin")
var RoleUser = Role("user")
var RoleGuest = Role("guest")
var RoleBanned = Role("banned")

func IsRole(ctx context.Context, role Role) bool {
	if user := ParseUserIdentity(ctx); user != nil {
		return user.Role == role
	}

	return false
}

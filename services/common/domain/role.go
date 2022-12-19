package domain

type Role string

var RoleAdmin = Role("admin")
var RoleUser = Role("user")
var RoleBanned = Role("banned")

type Gettable interface {
	Get(string) interface{}
}

func IsRole(ctx Gettable, role Role) bool {
	if session, ok := ctx.Get("session").(*SessionToken); ok {
		return session.Role == role
	}

	return false
}

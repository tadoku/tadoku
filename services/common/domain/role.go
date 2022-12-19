package domain

import (
	"github.com/labstack/echo/v4"
)

type Role string

var RoleAdmin = Role("admin")
var RoleUser = Role("user")
var RoleBanned = Role("banned")

func IsRole(ctx echo.Context, role Role) bool {
	if session, ok := ctx.Get("session").(*SessionToken); !ok {
		return session.Role == role
	}

	return false
}

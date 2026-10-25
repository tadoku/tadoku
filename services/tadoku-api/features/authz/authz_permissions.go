package authz

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

const appRoot = "root"

var Permissions = permissions.Define(
	"app",
	permissions.Relations{
		"admins": {permissions.Users},
		"banned": {permissions.Users},
	},
	permissions.Permits{
		"admin":     permissions.Related("admins"),
		"is_banned": permissions.Related("banned"),
	},
)

func adminPermission() permissions.Check {
	return Permissions.Check(appRoot, "admin")
}

func isBannedPermission() permissions.Check {
	return Permissions.Check(appRoot, "is_banned")
}

func adminRelationship(userID uuid.UUID) permissions.Relationship {
	return Permissions.UserRelationship(appRoot, "admins", userID)
}

func bannedRelationship(userID uuid.UUID) permissions.Relationship {
	return Permissions.UserRelationship(appRoot, "banned", userID)
}

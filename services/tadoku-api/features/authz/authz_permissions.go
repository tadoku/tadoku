package authz

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

const appRoot = "root"

var Permissions = permissions.MustDefine(
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

var (
	adminPermit    = Permissions.MustPermit("admin")
	isBannedPermit = Permissions.MustPermit("is_banned")
	adminsRelation = Permissions.MustUserRelation("admins")
	bannedRelation = Permissions.MustUserRelation("banned")
)

func adminPermission() permissions.Check {
	return adminPermit.On(appRoot)
}

func isBannedPermission() permissions.Check {
	return isBannedPermit.On(appRoot)
}

func adminRelationship(userID uuid.UUID) permissions.Relationship {
	return adminsRelation.On(appRoot, userID)
}

func bannedRelationship(userID uuid.UUID) permissions.Relationship {
	return bannedRelation.On(appRoot, userID)
}

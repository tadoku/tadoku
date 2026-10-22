package permissions

import "github.com/google/uuid"

const App ObjectType = "app"

var appRoot = object{objectType: App, id: "root"}

var model = Model{
	App: {
		Relations: Relations{
			"admins": {Users},
			"banned": {Users},
		},
		Permits: Permits{
			"admin":     Related("admins"),
			"is_banned": Related("banned"),
		},
	},
}

var (
	appAdmin  = Check{object: appRoot, permit: "admin"}
	appBanned = Check{object: appRoot, permit: "is_banned"}
)

func Admin(userID uuid.UUID) Relationship {
	return Relationship{object: appRoot, relation: "admins", subject: userSubject(userID)}
}

func Ban(userID uuid.UUID) Relationship {
	return Relationship{object: appRoot, relation: "banned", subject: userSubject(userID)}
}

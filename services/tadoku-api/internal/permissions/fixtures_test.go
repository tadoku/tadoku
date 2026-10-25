package permissions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const (
	Site    ObjectType = "site"
	Group   ObjectType = "group"
	Contest ObjectType = "contest"
	Log     ObjectType = "log"
)

const siteRoot = "root"

var (
	sites = MustDefine(
		Site,
		Relations{
			"admins": {Users},
			"banned": {Users},
		},
		Permits{"admin": Related("admins")},
	)
	groups = MustDefine(
		Group,
		Relations{"members": {Users, SetOf(Group, "members")}},
		nil,
	)
	contests = MustDefine(
		Contest,
		Relations{
			"owners":     {Users},
			"moderators": {Users, SetOf(Group, "members")},
		},
		Permits{
			"edit":     AnyOf(Related("owners"), PermitOn(Site, siteRoot, "admin")),
			"moderate": AnyOf(Allowed("edit"), Related("moderators")),
		},
	)
	logs = MustDefine(
		Log,
		Relations{"contests": {ObjectOf(Contest)}},
		Permits{"hide": AnyOf(PermitOn(Site, siteRoot, "admin"), Via("contests", "moderate"))},
	)
	testModel = mustModel(sites, groups, contests, logs)
	siteAdmin = sites.Check(siteRoot, "admin")
)

func mustModel(namespaces ...Namespace) Model {
	model, err := NewModel(namespaces...)
	if err != nil {
		panic(err)
	}
	return model
}

func contestObject(id uuid.UUID) object { return object{objectType: Contest, id: id.String()} }
func groupObject(id uuid.UUID) object   { return object{objectType: Group, id: id.String()} }

func groupMembers(groupID uuid.UUID) subjectSet {
	return subjectSet{object: groupObject(groupID), relation: "members"}
}

func productionContext(t *testing.T) context.Context {
	t.Helper()

	return tenant.WithKey(t.Context(), tenant.Production())
}

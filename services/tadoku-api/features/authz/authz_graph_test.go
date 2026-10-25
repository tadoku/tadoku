package authz

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"pgregory.net/rapid"
)

var propertySeeds = []uint64{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}

// graphPool is the small set of objects and users that generated relationships draw from, so that objects appear
// both as objects and as subjects and repeated draws collide.
type graphPool struct {
	users    []uuid.UUID
	contests []string
	groups   []string
	logs     []string
}

func newGraphPool() graphPool {
	pool := graphPool{
		contests: []string{"c0", "c1", "c2"},
		groups:   []string{"g0", "g1", "g2"},
		logs:     []string{"l0", "l1", "l2"},
	}
	for i := range 4 {
		pool.users = append(pool.users, uuid.MustParse(fmt.Sprintf("00000000-0000-4000-8000-%012d", i)))
	}
	return pool
}

func (p graphPool) objects() []permissions.Object {
	objects := []permissions.Object{sites.Object(siteRoot)}
	for _, id := range p.contests {
		objects = append(objects, contests.Object(id))
	}
	for _, id := range p.groups {
		objects = append(objects, groups.Object(id))
	}
	for _, id := range p.logs {
		objects = append(objects, logs.Object(id))
	}
	return objects
}

// relationship generates relationships of the test model, of every subject kind.
func (p graphPool) relationship() *rapid.Generator[permissions.Relationship] {
	user, contest := rapid.SampledFrom(p.users), rapid.SampledFrom(p.contests)
	group, log := rapid.SampledFrom(p.groups), rapid.SampledFrom(p.logs)
	return rapid.Custom(func(t *rapid.T) permissions.Relationship {
		switch rapid.IntRange(0, 7).Draw(t, "kind") {
		case 0:
			return siteAdmins.On(siteRoot, user.Draw(t, "user"))
		case 1:
			return siteBans.On(siteRoot, user.Draw(t, "user"))
		case 2:
			return ownersRelation.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 3:
			return moderatorUsers.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 4:
			return moderatorGroups.On(contest.Draw(t, "contest"), group.Draw(t, "group"))
		case 5:
			return memberUsers.On(group.Draw(t, "group"), user.Draw(t, "user"))
		case 6:
			return memberGroups.On(group.Draw(t, "group"), group.Draw(t, "subgroup"))
		default:
			return contestsRelation.On(log.Draw(t, "log"), contest.Draw(t, "contest"))
		}
	})
}

// invalidRelationships are built with handles but fail validation against testModel.
func (p graphPool) invalidRelationships() []permissions.Relationship {
	return []permissions.Relationship{
		ownersRelation.On("", p.users[0]),
		memberGroups.On(p.groups[0], ""),
		refParents.On(p.groups[0], p.groups[1]),
	}
}

// batch prints as its rows, in order, in rapid's counterexamples.
type batch []permissions.Relationship

func (b batch) GoString() string {
	formatted := make([]string, len(b))
	for i, relationship := range b {
		formatted[i] = row(relationship)
	}
	return "[" + strings.Join(formatted, " ") + "]"
}

func batchOf(relationship *rapid.Generator[permissions.Relationship], maxLen int) *rapid.Generator[batch] {
	return rapid.Map(rapid.SliceOfN(relationship, 0, maxLen), func(relationships []permissions.Relationship) batch {
		return relationships
	})
}

// writes generates Grant and Revoke batches, some with one invalid relationship at any position.
func (p graphPool) writes() *rapid.Generator[batch] {
	valid, invalid := batchOf(p.relationship(), 8), rapid.SampledFrom(p.invalidRelationships())
	return rapid.Custom(func(t *rapid.T) batch {
		relationships := valid.Draw(t, "valid")
		if rapid.IntRange(0, 3).Draw(t, "invalid") == 3 {
			at := rapid.IntRange(0, len(relationships)).Draw(t, "at")
			relationships = slices.Insert(relationships, at, invalid.Draw(t, "relationship"))
		}
		return relationships
	})
}

var testNamespaces = []permissions.Namespace{sites, groups, contests, logs}

// testRelation names a relation handle of testModel, for Read filters.
type testRelation struct {
	handle     permissions.Relation
	objectType permissions.ObjectType
	name       string
}

var testRelations = []testRelation{
	{siteAdmins, siteType, "admins"},
	{siteBans, siteType, "banned"},
	{ownersRelation, contestType, "owners"},
	{moderatorUsers, contestType, "moderators"},
	{moderatorGroups, contestType, "moderators"},
	{memberUsers, groupType, "members"},
	{memberGroups, groupType, "members"},
	{contestsRelation, logType, "contests"},
}

// read is a valid Read filter and the relationships it selects.
type read struct {
	filter  permissions.Filter
	selects func(permissions.Relationship) bool
	name    string
}

func (r read) GoString() string { return r.name }

// reads generates filters on a pool object or for a pool user, some narrowed to a namespace, a relation or both.
func (p graphPool) reads() *rapid.Generator[read] {
	objects, users := rapid.SampledFrom(p.objects()), rapid.SampledFrom(p.users)
	return rapid.Custom(func(t *rapid.T) read {
		var r read
		var objectType permissions.ObjectType
		if rapid.Bool().Draw(t, "for a user") {
			user := users.Draw(t, "user")
			r = read{
				filter:  permissions.ForUser(user),
				selects: storesUser(user),
				name:    fmt.Sprintf("ForUser(%s)", user),
			}
			if rapid.Bool().Draw(t, "in a namespace") {
				namespace := rapid.SampledFrom(testNamespaces).Draw(t, "namespace")
				objectType = namespace.Object(siteRoot).Type()
				r.filter, r.name = r.filter.In(namespace), fmt.Sprintf("%s.In(%s)", r.name, objectType)
			}
		} else {
			object := objects.Draw(t, "object")
			objectType = object.Type()
			r = read{
				filter:  permissions.OnObject(object),
				selects: func(x permissions.Relationship) bool { return x.Object() == object },
				name:    fmt.Sprintf("OnObject(%s:%s)", objectType, object.ID()),
			}
		}

		relations := testRelations
		if objectType != "" {
			relations = slices.DeleteFunc(slices.Clone(relations), func(relation testRelation) bool {
				return relation.objectType != objectType
			})
		}
		var relation testRelation
		if rapid.Bool().Draw(t, "one relation") {
			relation = rapid.SampledFrom(relations).Draw(t, "relation")
			r.filter, r.name = r.filter.Relation(relation.handle), fmt.Sprintf("%s.Relation(%s)", r.name, relation.name)
		}

		anchored := r.selects
		r.selects = func(x permissions.Relationship) bool {
			return anchored(x) &&
				(objectType == "" || x.Object().Type() == objectType) &&
				(relation.name == "" || x.Object().Type() == relation.objectType && x.Relation() == relation.name)
		}
		return r
	})
}

func storesUser(user uuid.UUID) func(permissions.Relationship) bool {
	return func(relationship permissions.Relationship) bool {
		subject, ok := relationship.User()
		return ok && subject == user
	}
}

// row formats a relationship as storedRelationships reads it back.
func row(relationship permissions.Relationship) string {
	object := relationship.Object()
	subject := ""
	if user, ok := relationship.User(); ok {
		subject = permissions.UserSubjectType + ":" + user.String() + "#"
	}
	if target, ok := relationship.Target(); ok {
		subject = string(target.Type()) + ":" + target.ID() + "#"
	}
	if set, relation, ok := relationship.Set(); ok {
		subject = string(set.Type()) + ":" + set.ID() + "#" + relation
	}
	return string(object.Type()) + ":" + object.ID() + "#" + relationship.Relation() + "@" + subject
}

// rows returns the distinct rows of relationships, sorted as storedRelationships returns them.
func rows(relationships []permissions.Relationship) []string {
	formatted := make([]string, len(relationships))
	for i, relationship := range relationships {
		formatted[i] = row(relationship)
	}
	slices.Sort(formatted)
	return slices.Compact(formatted)
}

// rowDifference lists the rows missing from got and the extra rows in it.
func rowDifference(got, want []string) string {
	var missing, extra []string
	for _, row := range want {
		if !slices.Contains(got, row) {
			missing = append(missing, row)
		}
	}
	for _, row := range got {
		if !slices.Contains(want, row) {
			extra = append(extra, row)
		}
	}
	return fmt.Sprintf("missing %q, extra %q", missing, extra)
}

func propertyFailure(t *testing.T, seed uint64, graph []permissions.Relationship, format string, args ...any) {
	t.Helper()

	t.Errorf("seed %d: %s\ngraph:\n  %s", seed, fmt.Sprintf(format, args...), strings.Join(rows(graph), "\n  "))
}

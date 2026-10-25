package authz

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

var propertySeeds = []uint64{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233}

// graphPool is the small set of objects and users a random graph draws from, so that objects appear both as
// objects and as subjects and repeated draws collide.
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

// randomRelationship draws one relationship of the test model, of any subject kind.
func (p graphPool) randomRelationship(r *rand.Rand) permissions.Relationship {
	user := p.users[r.IntN(len(p.users))]
	contest, group := p.contests[r.IntN(len(p.contests))], p.groups[r.IntN(len(p.groups))]
	switch r.IntN(8) {
	case 0:
		return siteAdmins.On(siteRoot, user)
	case 1:
		return siteBans.On(siteRoot, user)
	case 2:
		return ownersRelation.On(contest, user)
	case 3:
		return moderatorUsers.On(contest, user)
	case 4:
		return moderatorGroups.On(contest, group)
	case 5:
		return memberUsers.On(group, user)
	case 6:
		return memberGroups.On(group, p.groups[r.IntN(len(p.groups))])
	default:
		return contestsRelation.On(p.logs[r.IntN(len(p.logs))], contest)
	}
}

func (p graphPool) randomRelationships(r *rand.Rand, count int) []permissions.Relationship {
	relationships := make([]permissions.Relationship, count)
	for i := range relationships {
		relationships[i] = p.randomRelationship(r)
	}
	return relationships
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

func keep(
	relationships []permissions.Relationship,
	kept func(permissions.Relationship) bool,
) []permissions.Relationship {
	var selected []permissions.Relationship
	for _, relationship := range relationships {
		if kept(relationship) {
			selected = append(selected, relationship)
		}
	}
	return selected
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

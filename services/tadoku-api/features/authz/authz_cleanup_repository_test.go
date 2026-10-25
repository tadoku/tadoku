package authz

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestRevokeObject(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/revoke-object-0000000a")
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	user, other := uuid.New(), uuid.New()
	shared, otherContest, otherGroup, logID := "x1", "c2", "g2", "l1"
	contest, group := contests.Object(shared), groups.Object(shared)
	onContest := []permissions.Relationship{
		ownersRelation.On(shared, user),
		moderatorUsers.On(shared, other),
		moderatorGroups.On(shared, otherGroup),
		contestsRelation.On(logID, shared),
	}
	onGroup := []permissions.Relationship{
		memberUsers.On(shared, user),
		memberGroups.On(shared, otherGroup),
		memberGroups.On(otherGroup, shared),
		moderatorGroups.On(otherContest, shared),
	}
	unrelated := []permissions.Relationship{
		ownersRelation.On(otherContest, user),
		contestsRelation.On(logID, otherContest),
		memberUsers.On(otherGroup, other),
	}
	everything := slices.Concat(onContest, onGroup, unrelated)
	grantAll(t, ctx, writer, everything...)
	grantAll(t, branchCtx, writer, everything...)

	steps := []struct {
		name   string
		object permissions.Object
		want   []string
	}{
		{"contest with a group of the same ID", contest, rows(slices.Concat(onGroup, unrelated))},
		{"group named as a subject set and by a subgroup", group, rows(unrelated)},
	}
	for _, step := range steps {
		statements.take()

		err := writer.RevokeObject(ctx, step.object)

		sent := statements.take()
		if err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if stored := storedRelationships(t, db, tenant.Production().String()); !slices.Equal(stored, step.want) {
			t.Errorf("%s: stored rows %s", step.name, rowDifference(stored, step.want))
		}
		if len(sent) != 1 {
			t.Errorf("%s: sent %d statements, want 1: %q", step.name, len(sent), sent)
		}
	}
	if stored := storedRelationships(t, db, "e2e/revoke-object-0000000a"); !slices.Equal(stored, rows(everything)) {
		t.Errorf("other tenant: stored rows %s", rowDifference(stored, rows(everything)))
	}

	t.Run("invalid objects", func(t *testing.T) {
		badges := permissions.MustDefine("badge", permissions.Relations{"holders": {permissions.Users}}, nil)
		for name, object := range map[string]permissions.Object{
			"zero value":                       {},
			"missing ID":                       contests.Object(""),
			"namespace missing from the model": badges.Object("b1"),
		} {
			statements.take()

			err := writer.RevokeObject(ctx, object)

			if err == nil {
				t.Errorf("%s: RevokeObject accepted an invalid object", name)
			}
			if sent := statements.take(); len(sent) != 0 {
				t.Errorf("%s: sent %d statements, want none", name, len(sent))
			}
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		if err := writer.RevokeObject(t.Context(), contest); errx.KindOf(err) != errx.Unavailable {
			t.Errorf("RevokeObject without tenant error = %v, want unavailable", err)
		}
	})
}

func TestRevokeUser(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/revoke-user-0000000a")
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	user, other := uuid.New(), uuid.New()
	namedLikeUser := user.String()
	ofUser := []permissions.Relationship{
		siteAdminRelationship(user),
		ownersRelation.On("c1", user),
		memberUsers.On("g1", user),
	}
	kept := []permissions.Relationship{
		siteAdminRelationship(other),
		memberUsers.On(namedLikeUser, other),
		memberGroups.On("g1", namedLikeUser),
		moderatorGroups.On("c1", namedLikeUser),
	}
	everything := slices.Concat(ofUser, kept)
	grantAll(t, ctx, writer, everything...)
	grantAll(t, branchCtx, writer, everything...)
	statements.take()

	err := writer.RevokeUser(ctx, user)

	sent := statements.take()
	if err != nil {
		t.Fatal(err)
	}
	if stored := storedRelationships(t, db, tenant.Production().String()); !slices.Equal(stored, rows(kept)) {
		t.Errorf("stored rows %s", rowDifference(stored, rows(kept)))
	}
	if stored := storedRelationships(t, db, "e2e/revoke-user-0000000a"); !slices.Equal(stored, rows(everything)) {
		t.Errorf("other tenant: stored rows %s", rowDifference(stored, rows(everything)))
	}
	if len(sent) != 1 {
		t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
	}
	if err := writer.RevokeUser(t.Context(), user); errx.KindOf(err) != errx.Unavailable {
		t.Errorf("RevokeUser without tenant error = %v, want unavailable", err)
	}
}

// TestCleanupProperties checks, for random graphs, that Read returns the rows on an object and for a user, that no
// row names the object after RevokeObject or the user after RevokeUser, that every other row stays, and that each
// call is one statement.
func TestCleanupProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	graphs := newGraphPool()
	objects := graphs.objects()

	for _, seed := range propertySeeds {
		key := fmt.Sprintf("e2e/cleanup-properties-%d", seed)
		ctx := registerTestTenant(t, db, key)
		r := rand.New(rand.NewPCG(seed, 0))
		graph := graphs.randomRelationships(r, 1+r.IntN(30))
		grantAll(t, ctx, writer, graph...)
		object, user := objects[r.IntN(len(objects))], graphs.users[r.IntN(len(graphs.users))]
		storesUser := func(r permissions.Relationship) bool {
			subject, ok := r.User()
			return ok && subject == user
		}
		withoutObject := keep(graph, func(r permissions.Relationship) bool { return !namesObject(r, object) })
		withoutEither := keep(withoutObject, func(r permissions.Relationship) bool { return !storesUser(r) })
		check := func(step string, revoke func(context.Context) error, want []string) {
			t.Helper()
			statements.take()

			err := revoke(ctx)

			sent := statements.take()
			if err != nil {
				propertyFailure(t, seed, graph, "%s: %v", step, err)
			}
			if stored := storedRelationships(t, db, key); !slices.Equal(stored, want) {
				propertyFailure(t, seed, graph, "%s: stored rows %s", step, rowDifference(stored, want))
			}
			if len(sent) != 1 {
				propertyFailure(t, seed, graph, "%s: sent %d statements, want 1", step, len(sent))
			}
		}

		for name, read := range map[string]struct {
			filter permissions.Filter
			keep   func(permissions.Relationship) bool
		}{
			"read on the object": {permissions.OnObject(object), func(r permissions.Relationship) bool {
				return r.Object() == object
			}},
			"read for the user": {permissions.ForUser(user), storesUser},
		} {
			statements.take()

			got, err := writer.Read(ctx, read.filter)

			want := rows(keep(graph, read.keep))
			if err != nil || !slices.Equal(rows(got), want) {
				propertyFailure(t, seed, graph, "%s: %v, rows %s", name, err, rowDifference(rows(got), want))
			}
			if sent := statements.take(); len(sent) != 1 {
				propertyFailure(t, seed, graph, "%s: sent %d statements, want 1", name, len(sent))
			}
		}
		check(
			fmt.Sprintf("revoke object %s:%s", object.Type(), object.ID()),
			func(ctx context.Context) error { return writer.RevokeObject(ctx, object) },
			rows(withoutObject),
		)
		check(
			fmt.Sprintf("revoke user %s", user),
			func(ctx context.Context) error { return writer.RevokeUser(ctx, user) },
			rows(withoutEither),
		)
	}
}

func namesObject(relationship permissions.Relationship, object permissions.Object) bool {
	target, _ := relationship.Target()
	set, _, _ := relationship.Set()
	return relationship.Object() == object || target == object || set == object
}

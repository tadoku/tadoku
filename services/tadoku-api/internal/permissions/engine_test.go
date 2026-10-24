package permissions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCheckFailsClosed(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	cycleContest, first, second := uuid.New(), uuid.New(), uuid.New()
	contestID, moderator := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		NewRelationships(db.AppPool, testModel),
		contestModeratorGroup(cycleContest, first),
		subgroup(first, second),
		subgroup(second, first),
		contestModerator(contestID, moderator),
	)
	working := newStore(db.AppPool)

	closedPool, err := pgxpool.NewWithConfig(ctx, db.AppPool.Config())
	if err != nil {
		t.Fatal(err)
	}
	closedPool.Close()

	cancelled, cancel := context.WithCancel(ctx)
	cancel()

	unknownPermit := Check{object: contestObject(contestID), permit: "delete"}
	unknownType := Check{object: object{objectType: "badge", id: "1"}, permit: "view"}

	cases := []struct {
		name  string
		ctx   context.Context
		store store
		check Check
		user  uuid.UUID
		want  error
	}{
		{"group cycle", ctx, working, moderateContest(cycleContest), uuid.New(), errMaxDepth},
		{"storage error", ctx, newStore(closedPool), moderateContest(contestID), moderator, nil},
		{"unknown permit", ctx, working, unknownPermit, moderator, nil},
		{"unknown object type", ctx, working, unknownType, moderator, nil},
		{"cancelled context", cancelled, working, moderateContest(contestID), moderator, context.Canceled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := testModel.check(tc.ctx, tc.store, tc.check, tc.user)

			if err == nil {
				t.Fatal("check returned no error")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
			if allowed {
				t.Error("check failed open")
			}
		})
	}
}

func TestCombinators(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	suspensions, err := NewModel(groups, Define(
		Contest,
		Relations{
			"moderators": {Users},
			"suspended":  {Users, SetOf(Group, "members")},
			"verified":   {Users},
		},
		Permits{
			"moderate": Except(Related("moderators"), Related("suspended")),
			"review":   AllOf(Related("moderators"), Related("verified")),
		},
	))
	if err != nil {
		t.Fatal(err)
	}

	db := newTestDatabase(t)
	contestID, active, suspended, verified := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	cycleContest, cycleModerator, first, second := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contest := func(id uuid.UUID, relation string, subject subject) Relationship {
		return Relationship{object: contestObject(id), relation: relation, subject: subject}
	}
	grantAll(
		t,
		ctx,
		NewRelationships(db.AppPool, suspensions),
		contest(contestID, "moderators", userSubject(active)),
		contest(contestID, "moderators", userSubject(suspended)),
		contest(contestID, "suspended", userSubject(suspended)),
		contest(contestID, "moderators", userSubject(verified)),
		contest(contestID, "verified", userSubject(verified)),
		contest(cycleContest, "moderators", userSubject(cycleModerator)),
		contest(cycleContest, "suspended", groupMembers(first)),
		subgroup(first, second),
		subgroup(second, first),
	)

	cases := []struct {
		name   string
		object object
		permit string
		user   uuid.UUID
		want   bool
	}{
		{"active moderator moderates", contestObject(contestID), "moderate", active, true},
		{"suspended moderator cannot moderate", contestObject(contestID), "moderate", suspended, false},
		{"verified moderator reviews", contestObject(contestID), "review", verified, true},
		{"unverified moderator cannot review", contestObject(contestID), "review", active, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := Check{object: tc.object, permit: tc.permit}

			allowed, err := suspensions.check(ctx, newStore(db.AppPool), check, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}

	t.Run("error in the subtracted rule", func(t *testing.T) {
		check := Check{object: contestObject(cycleContest), permit: "moderate"}

		allowed, err := suspensions.check(ctx, newStore(db.AppPool), check, cycleModerator)

		if !errors.Is(err, errMaxDepth) {
			t.Errorf("error = %v, want %v", err, errMaxDepth)
		}
		if allowed {
			t.Error("check failed open")
		}
	})
}

func TestModelValidation(t *testing.T) {
	t.Parallel()

	owners := Relations{"owners": {Users}}
	cases := []struct {
		name       string
		namespaces []Namespace
	}{
		{"unknown relation", []Namespace{Define(Contest, nil, Permits{"edit": Related("owners")})}},
		{"unknown permit", []Namespace{Define(Contest, nil, Permits{"edit": Allowed("moderate")})}},
		{"relation and permit share a name", []Namespace{
			Define(Contest, Relations{"edit": {Users}}, Permits{"edit": Related("edit")}),
		}},
		{"via over users", []Namespace{Define(Contest, owners, Permits{"edit": Via("owners", "edit")})}},
		{"related over objects", []Namespace{
			Define(Contest, nil, nil),
			Define(Log, Relations{"contests": {ObjectOf(Contest)}}, Permits{"hide": Related("contests")}),
		}},
		{"via to a missing permit", []Namespace{
			Define(Contest, nil, nil),
			Define(Log, Relations{"contests": {ObjectOf(Contest)}}, Permits{"hide": Via("contests", "moderate")}),
		}},
		{"permit on an unknown object type", []Namespace{
			Define(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"permit on a missing permit", []Namespace{
			Define(Site, nil, nil),
			Define(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"permit on an object without an ID", []Namespace{
			sites,
			Define(Contest, nil, Permits{"edit": PermitOn(Site, "", "admin")}),
		}},
		{"single-rule combinator", []Namespace{Define(Contest, owners, Permits{"edit": AnyOf(Related("owners"))})}},
		{"unknown subject type", []Namespace{
			Define(Contest, Relations{"moderators": {SetOf(Group, "members")}}, nil),
		}},
		{"object type reserved for users", []Namespace{Define("user", Relations{"friends": {Users}}, nil)}},
		{"object type without a name", []Namespace{Define("", owners, nil)}},
		{"object type defined twice", []Namespace{Define(Contest, owners, nil), Define(Contest, nil, nil)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewModel(tc.namespaces...); err == nil {
				t.Error("invalid model passed validation")
			}
		})
	}
}

func TestRelationshipValidation(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	pool, statements := newLoggedAppPool(t, db)
	writer := NewRelationships(pool, testModel)
	contest, user := contestObject(uuid.New()), userSubject(uuid.New())

	cases := []struct {
		name         string
		relationship Relationship
	}{
		{"group as owner", Relationship{object: contest, relation: "owners", subject: groupMembers(uuid.New())}},
		{"unknown relation", Relationship{object: contest, relation: "judges", subject: user}},
		{"missing subject", Relationship{object: contest, relation: "owners"}},
		{"missing object ID", Relationship{object: object{objectType: Contest}, relation: "owners", subject: user}},
		{"zero value", Relationship{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			grantErr := writer.Grant(ctx, tc.relationship)
			revokeErr := writer.Revoke(ctx, tc.relationship)
			_, storedErr := writer.Stored(ctx, []Relationship{tc.relationship})

			if grantErr == nil {
				t.Error("grant accepted an invalid relationship")
			}
			if revokeErr == nil {
				t.Error("revoke accepted an invalid relationship")
			}
			if storedErr == nil {
				t.Error("stored accepted an invalid relationship")
			}
			if sent := statements.take(); len(sent) != 0 {
				t.Errorf("sent %d statements before rejecting: %q", len(sent), sent)
			}
		})
	}
}

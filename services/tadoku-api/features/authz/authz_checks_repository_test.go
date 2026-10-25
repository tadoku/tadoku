package authz

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

const maxDepthError = "permission check exceeded the maximum depth"

func TestChecksStayInTheirTenant(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000a")
	repository := NewRepository(db.AppPool)
	writer := permissions.NewRelationships(repository, testModel)
	productionAdmin, branchAdmin, stranger := uuid.New(), uuid.New(), uuid.New()
	grantAll(t, prodCtx, writer, siteAdminRelationship(productionAdmin))
	grantAll(t, branchCtx, writer, siteAdminRelationship(branchAdmin))
	checker := permissions.NewChecker(repository, testModel)

	cases := []struct {
		name string
		ctx  context.Context
		user uuid.UUID
		want bool
	}{
		{"production admin is admin in production", prodCtx, productionAdmin, true},
		{"stranger is not admin", prodCtx, stranger, false},
		{"production admin is not admin on a test tenant", branchCtx, productionAdmin, false},
		{"test tenant admin is admin in its tenant", branchCtx, branchAdmin, true},
		{"test tenant admin is not admin in production", prodCtx, branchAdmin, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := checker.AllowedFor(tc.ctx, siteAdmin, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}
}

func TestRelationshipGraphChecks(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	admin, owner, moderator, teamMember, nestedMember, stranger := uuid.New(), uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), uuid.New()
	contestA, contestB := uuid.New(), uuid.New()
	team, nestedTeam := uuid.New(), uuid.New()
	logA, logB := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, testModel),
		siteAdminRelationship(admin),
		contestOwner(contestA, owner),
		contestModerator(contestA, moderator),
		groupMember(team, teamMember),
		groupMember(nestedTeam, nestedMember),
	)
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, testModel),
		contestModeratorGroup(contestA, team),
		subgroup(team, nestedTeam),
		logCountsFor(logA, contestA),
		logCountsFor(logB, contestB),
	)
	checker := permissions.NewChecker(repository, testModel)

	cases := []struct {
		name  string
		check permissions.Check
		user  uuid.UUID
		want  bool
	}{
		{"site admin edits any contest", editContest(contestB), admin, true},
		{"owner edits own contest", editContest(contestA), owner, true},
		{"owner cannot edit another contest", editContest(contestB), owner, false},
		{"owner moderates through edit", moderateContest(contestA), owner, true},
		{"moderator moderates", moderateContest(contestA), moderator, true},
		{"moderator cannot edit", editContest(contestA), moderator, false},
		{"team member moderates through group", moderateContest(contestA), teamMember, true},
		{"nested team member moderates", moderateContest(contestA), nestedMember, true},
		{"team member cannot moderate another contest", moderateContest(contestB), teamMember, false},
		{"moderator hides log in their contest", hideLog(logA), moderator, true},
		{"nested team member hides log in their contest", hideLog(logA), nestedMember, true},
		{"moderator cannot hide log in another contest", hideLog(logB), moderator, false},
		{"site admin hides any log", hideLog(logB), admin, true},
		{"stranger cannot moderate", moderateContest(contestA), stranger, false},
		{"stranger cannot hide", hideLog(logA), stranger, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := checker.AllowedFor(ctx, tc.check, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}
}

func TestCheckFailsClosed(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	cycleContest, first, second := uuid.New(), uuid.New(), uuid.New()
	contestID, moderator := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, testModel),
		contestModeratorGroup(cycleContest, first),
		subgroup(first, second),
		subgroup(second, first),
		contestModerator(contestID, moderator),
	)
	working := permissions.NewChecker(repository, testModel)
	broken := permissions.NewChecker(NewRepository(closedPool(t, db)), testModel)

	cancelled, cancel := context.WithCancel(ctx)
	cancel()

	badges := permissions.MustDefine(
		"badge",
		permissions.Relations{"holders": {permissions.Users}},
		permissions.Permits{"view": permissions.Related("holders")},
	)
	unknownType := badges.MustPermit("view").On("1")

	cases := []struct {
		name     string
		ctx      context.Context
		checker  *permissions.Checker
		check    permissions.Check
		user     uuid.UUID
		want     error
		wantText string
	}{
		{"group cycle", ctx, working, moderateContest(cycleContest), uuid.New(), nil, maxDepthError},
		{"storage error", ctx, broken, moderateContest(contestID), moderator, nil, ""},
		{"unknown permit", ctx, working, permissions.Check{}, moderator, nil, ""},
		{"unknown object type", ctx, working, unknownType, moderator, nil, ""},
		{"cancelled context", cancelled, working, moderateContest(contestID), moderator, context.Canceled, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := tc.checker.AllowedFor(tc.ctx, tc.check, tc.user)

			if err == nil {
				t.Fatal("check returned no error")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.wantText) {
				t.Errorf("error = %v, want %q", err, tc.wantText)
			}
			if allowed {
				t.Error("check failed open")
			}
		})
	}

	t.Run("cancelled after the lookup", func(t *testing.T) {
		admin := uuid.New()
		grantAll(t, ctx, permissions.NewRelationships(repository, testModel), siteAdminRelationship(admin))
		pool, statements := newLoggedPool(t, db)
		lookupCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		statements.onBatchEnd = cancel
		checker := permissions.NewChecker(NewRepository(pool), testModel)

		allowed, err := checker.AllowedFor(lookupCtx, siteAdmin, admin)

		if allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, context.Canceled) {
			t.Errorf("AllowedFor = %t, %v; want false and unavailable wrapping context.Canceled", allowed, err)
		}
	})
}

func TestCombinators(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	suspendable := permissions.MustDefine(
		contestType,
		permissions.Relations{
			"moderators": {permissions.Users},
			"suspended":  {permissions.Users, permissions.SetOf(groupType, "members")},
			"verified":   {permissions.Users},
		},
		permissions.Permits{
			"moderate": permissions.Except(permissions.Related("moderators"), permissions.Related("suspended")),
			"review":   permissions.AllOf(permissions.Related("moderators"), permissions.Related("verified")),
		},
	)
	suspensions, err := permissions.NewModel(groups, suspendable)
	if err != nil {
		t.Fatal(err)
	}

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	contestID, active, suspended, verified := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	cycleContest, cycleModerator, first, second := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	moderators, suspendedUsers := suspendable.MustUserRelation("moderators"), suspendable.MustUserRelation("suspended")
	verifiedUsers, suspendedGroups := suspendable.MustUserRelation("verified"), suspendable.MustSetRelation("suspended")
	moderate, review := suspendable.MustPermit("moderate"), suspendable.MustPermit("review")
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, suspensions),
		moderators.On(contestID.String(), active),
		moderators.On(contestID.String(), suspended),
		suspendedUsers.On(contestID.String(), suspended),
		moderators.On(contestID.String(), verified),
		verifiedUsers.On(contestID.String(), verified),
		moderators.On(cycleContest.String(), cycleModerator),
		suspendedGroups.On(cycleContest.String(), first.String()),
		subgroup(first, second),
		subgroup(second, first),
	)
	checker := permissions.NewChecker(repository, suspensions)

	cases := []struct {
		name    string
		contest uuid.UUID
		permit  permissions.Permit
		user    uuid.UUID
		want    bool
	}{
		{"active moderator moderates", contestID, moderate, active, true},
		{"suspended moderator cannot moderate", contestID, moderate, suspended, false},
		{"verified moderator reviews", contestID, review, verified, true},
		{"unverified moderator cannot review", contestID, review, active, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := tc.permit.On(tc.contest.String())

			allowed, err := checker.AllowedFor(ctx, check, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}

	t.Run("error in the subtracted rule", func(t *testing.T) {
		check := moderate.On(cycleContest.String())

		allowed, err := checker.AllowedFor(ctx, check, cycleModerator)

		if err == nil || !strings.Contains(err.Error(), maxDepthError) {
			t.Errorf("error = %v, want %q", err, maxDepthError)
		}
		if allowed {
			t.Error("check failed open")
		}
	})
}

func TestViaPastMaxDepthFailsClosed(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	nested := permissions.MustDefine(
		groupType,
		permissions.Relations{
			"members": {permissions.Users},
			"parents": {permissions.ObjectOf(groupType)},
		},
		permissions.Permits{
			"hop0":     permissions.Allowed("hop1"),
			"hop1":     permissions.Allowed("hop2"),
			"hop2":     permissions.Allowed("hop3"),
			"hop3":     permissions.Allowed("hop4"),
			"hop4":     permissions.Allowed("hop5"),
			"hop5":     permissions.Allowed("hop6"),
			"hop6":     permissions.Allowed("hop7"),
			"hop7":     permissions.Except(permissions.Related("members"), permissions.Allowed("excluded")),
			"excluded": permissions.Via("parents", "blocked"),
			"blocked":  permissions.Related("members"),
		},
	)
	model, err := permissions.NewModel(nested)
	if err != nil {
		t.Fatal(err)
	}

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	groupID, member := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, model),
		nested.MustUserRelation("members").On(groupID.String(), member),
	)
	checker := permissions.NewChecker(repository, model)

	allowed, err := checker.AllowedFor(ctx, nested.MustPermit("hop0").On(groupID.String()), member)

	if err == nil || !strings.Contains(err.Error(), maxDepthError) {
		t.Errorf("error = %v, want %q", err, maxDepthError)
	}
	if allowed {
		t.Error("check failed open")
	}
}

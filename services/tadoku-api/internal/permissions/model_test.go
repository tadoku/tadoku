package permissions

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestModelIsValid(t *testing.T) {
	t.Parallel()

	if _, err := NewModel(sites, groups, contests, logs); err != nil {
		t.Errorf("test model: %v", err)
	}
}

func TestChecksStayInTheirTenant(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000a")
	writer := NewRelationships(db.AppPool, testModel)
	productionAdmin, branchAdmin, stranger := uuid.New(), uuid.New(), uuid.New()
	grantAll(t, prodCtx, writer, adminRelationship(productionAdmin))
	grantAll(t, branchCtx, writer, adminRelationship(branchAdmin))

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
			allowed, err := testModel.check(tc.ctx, newStore(db.AppPool), siteAdmin, tc.user)

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
	admin, owner, moderator, teamMember, nestedMember, stranger := uuid.New(), uuid.New(), uuid.New(),
		uuid.New(), uuid.New(), uuid.New()
	contestA, contestB := uuid.New(), uuid.New()
	team, nestedTeam := uuid.New(), uuid.New()
	logA, logB := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		NewRelationships(db.AppPool, testModel),
		adminRelationship(admin),
		contestOwner(contestA, owner),
		contestModerator(contestA, moderator),
		contestModeratorGroup(contestA, team),
		groupMember(team, teamMember),
		subgroup(team, nestedTeam),
		groupMember(nestedTeam, nestedMember),
		logCountsFor(logA, contestA),
		logCountsFor(logB, contestB),
	)

	cases := []struct {
		name  string
		check Check
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
			allowed, err := testModel.check(ctx, newStore(db.AppPool), tc.check, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}
}

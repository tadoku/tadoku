package permissions

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestModelIsValid(t *testing.T) {
	t.Parallel()

	if err := model.validate(); err != nil {
		t.Errorf("model: %v", err)
	}
	if err := testModel.validate(); err != nil {
		t.Errorf("test model: %v", err)
	}
}

func TestAppModelChecks(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/permissions-0000000a")
	grants := NewGrants(db.AppPool)
	admin, bannedUser, branchAdmin, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	grantAll(t, prodCtx, grants, Admin(admin), Ban(bannedUser))
	grantAll(t, branchCtx, grants, Admin(branchAdmin))

	cases := []struct {
		name  string
		ctx   context.Context
		check Check
		user  uuid.UUID
		want  bool
	}{
		{"admin is admin", prodCtx, appAdmin, admin, true},
		{"stranger is not admin", prodCtx, appAdmin, stranger, false},
		{"banned user is not admin", prodCtx, appAdmin, bannedUser, false},
		{"banned user is banned", prodCtx, appBanned, bannedUser, true},
		{"admin is not banned", prodCtx, appBanned, admin, false},
		{"production admin is not admin on a test tenant", branchCtx, appAdmin, admin, false},
		{"production ban does not reach a test tenant", branchCtx, appBanned, bannedUser, false},
		{"test tenant admin is admin in its tenant", branchCtx, appAdmin, branchAdmin, true},
		{"test tenant admin is not admin in production", prodCtx, appAdmin, branchAdmin, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := model.check(tc.ctx, store{db: db.AppPool}, tc.check, tc.user)

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
		&Grants{model: testModel, store: store{db: db.AppPool}},
		Admin(admin),
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
		{"tenant admin edits any contest", editContest(contestB), admin, true},
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
		{"tenant admin hides any log", hideLog(logB), admin, true},
		{"stranger cannot moderate", moderateContest(contestA), stranger, false},
		{"stranger cannot hide", hideLog(logA), stranger, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allowed, err := testModel.check(ctx, store{db: db.AppPool}, tc.check, tc.user)

			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.want {
				t.Errorf("allowed = %t, want %t", allowed, tc.want)
			}
		})
	}
}

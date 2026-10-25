package authz

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// TestCheckQueryBudget asserts the queries a check sends: at most one per rule level it reaches.
func TestCheckQueryBudget(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	trios := permissions.MustDefine(
		"trio",
		permissions.Relations{"a": {permissions.Users}, "b": {permissions.Users}, "c": {permissions.Users}},
		permissions.Permits{"any": permissions.AnyOf(
			permissions.Related("a"),
			permissions.Related("b"),
			permissions.Related("c"),
		)},
	)
	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	repository := NewRepository(pool)
	checker := permissions.NewChecker(repository, refModel)
	trioChecker := permissions.NewChecker(repository, mustModel(trios))
	admin, moderator := uuid.New(), uuid.New()
	relationships := []permissions.Relationship{refSiteAdmins.On(siteRoot, admin), refMemberUsers.On("g1", moderator)}
	var edits []permissions.Check
	for i := range 50 {
		logID, contestID := fmt.Sprintf("l%d", i), fmt.Sprintf("c%d", i%5)
		relationships = append(
			relationships,
			refLogContests.On(logID, contestID),
			refModeratorMembers.On(contestID, "g1"),
		)
		edits = append(edits, refLogEdit.On(logID))
	}
	grantAll(t, ctx, permissions.NewRelationships(repository, refModel), relationships...)

	cases := []struct {
		name        string
		check       func() error
		wantQueries int
	}{
		{"admin check", func() error {
			return wantAllowed(checker.AllowedFor(ctx, refSiteAdmin.On(siteRoot), admin))
		}, 1},
		{"any of three relations", func() error {
			return wantDenied(trioChecker.AllowedFor(ctx, trios.MustPermit("any").On("t1"), admin))
		}, 1},
		{"edit through a contest a group moderates", func() error {
			return wantAllowed(checker.AllowedFor(ctx, edits[0], moderator))
		}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statements.take()

			err := tc.check()

			if err != nil {
				t.Fatal(err)
			}
			if sent := statements.take(); len(sent) != tc.wantQueries {
				t.Errorf("sent %d queries, want %d", len(sent), tc.wantQueries)
			}
		})
	}
}

func wantAllowed(allowed bool, err error) error {
	if err == nil && !allowed {
		err = errors.New("check denied, want allowed")
	}
	return err
}

func wantDenied(allowed bool, err error) error {
	if err == nil && allowed {
		err = errors.New("check allowed, want denied")
	}
	return err
}

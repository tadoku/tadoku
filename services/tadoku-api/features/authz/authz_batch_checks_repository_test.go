package authz

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestBatchedChecksAgreeWithAllowedFor(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	checker := permissions.NewChecker(repository, refModel)
	admin, owner, moderator, member, parentMember := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	suspended, verified, stranger := uuid.New(), uuid.New(), uuid.New()
	users := []uuid.UUID{admin, owner, moderator, member, parentMember, suspended, verified, stranger}
	graph := []permissions.Relationship{
		refSiteAdmins.On(siteRoot, admin),
		refOwners.On("c1", owner),
		refModeratorUsers.On("c1", moderator),
		refModeratorUsers.On("c1", suspended),
		refModeratorUsers.On("c1", verified),
		refSuspendedUsers.On("c1", suspended),
		refVerifiedUsers.On("c1", verified),
		refModeratorMembers.On("c1", "g1"),
		refMemberGroups.On("g1", "g2"),
		refMemberUsers.On("g2", member),
		refModeratorPermits.On("c1", "g3"),
		refParents.On("g3", "g4"),
		refMemberUsers.On("g4", parentMember),
		refLogOwners.On("l1", owner),
		refLogContests.On("l1", "c1"),
		refLogContests.On("l2", "c2"),
	}
	grantAll(t, ctx, permissions.NewRelationships(repository, refModel), graph...)
	questions := refQuestions(graphPool{
		users:    users,
		groups:   []string{"g1", "g2", "g3", "g4"},
		contests: []string{"c1", "c2"},
		logs:     []string{"l1", "l2"},
	})
	ref := newReference(graph)
	single := map[permissions.Check]map[uuid.UUID]bool{}
	for _, question := range questions {
		single[question.check] = map[uuid.UUID]bool{}
		for _, user := range users {
			allowed, err := checker.AllowedFor(ctx, question.check, user)
			if err != nil {
				t.Fatal(err)
			}
			if want := ref.check(question.object, question.permit, user); checkVerdict(allowed, nil) != want {
				t.Errorf("AllowedFor(%s#%s, user %d) = %t, reference %s",
					question.object.ID(), question.permit, slices.Index(users, user), allowed, want)
			}
			single[question.check][user] = allowed
		}
	}

	examples := []struct {
		name  string
		check permissions.Check
		user  uuid.UUID
		want  bool
	}{
		{"a nested group member moderates", refModerate.On("c1"), member, true},
		{"a member through a parent group moderates", refModerate.On("c1"), parentMember, true},
		{"a suspended moderator does not moderate", refModerate.On("c1"), suspended, false},
		{"a verified moderator hides a log", refLogHide.On("l1"), verified, true},
		{"an unverified moderator cannot hide it", refLogHide.On("l1"), moderator, false},
		{"an administrator hides any log", refLogHide.On("l2"), admin, true},
		{"a group member edits a log of the contest", refLogEdit.On("l1"), member, true},
		{"a stranger cannot edit", refLogEdit.On("l1"), stranger, false},
	}
	for _, example := range examples {
		if single[example.check][example.user] != example.want {
			t.Errorf("%s: AllowedFor = %t, want %t", example.name, !example.want, example.want)
		}
	}

	t.Run("AllowedEach", func(t *testing.T) {
		checks := make([]permissions.Check, len(questions))
		for i, question := range questions {
			checks[i] = question.check
		}
		for _, user := range users {
			allowed, err := checker.AllowedEach(signedIn(ctx, user.String()), checks...)

			if err != nil {
				t.Fatal(err)
			}
			for _, check := range checks {
				if allowed[check] != single[check][user] {
					t.Errorf("user %d: AllowedEach %+v = %t, AllowedFor %t",
						slices.Index(users, user), check, allowed[check], single[check][user])
				}
			}
		}
	})

	t.Run("AllowedForEach", func(t *testing.T) {
		for _, question := range questions {
			allowed, err := checker.AllowedForEach(ctx, question.check, users)

			if err != nil {
				t.Fatal(err)
			}
			for _, user := range users {
				if allowed[user] != single[question.check][user] {
					t.Errorf("%s#%s: AllowedForEach user %d = %t, AllowedFor %t", question.object.ID(),
						question.permit, slices.Index(users, user), allowed[user], single[question.check][user])
				}
			}
		}
	})
}

func TestBatchedChecksFailClosed(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	moderator, stranger := uuid.New(), uuid.New()
	grantAll(
		t,
		ctx,
		permissions.NewRelationships(repository, refModel),
		refSiteAdmins.On(siteRoot, moderator),
		refModeratorMembers.On("cycle", "g1"),
		refMemberGroups.On("g1", "g2"),
		refMemberGroups.On("g2", "g1"),
		refModeratorUsers.On("suspends", moderator),
		refSuspendedGroups.On("suspends", "p1"),
		refParents.On("p1", "p2"),
		refParents.On("p2", "p1"),
		refLogContests.On("linked", "cycle"),
	)
	working := permissions.NewChecker(repository, refModel)
	broken := permissions.NewChecker(NewRepository(closedPool(t, db)), refModel)
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	admin := refSiteAdmin.On(siteRoot)
	users := []uuid.UUID{moderator, stranger}

	cases := []struct {
		name     string
		ctx      context.Context
		checker  *permissions.Checker
		actor    uuid.UUID
		checks   []permissions.Check
		cause    error
		wantText string
	}{
		{"storage error", ctx, broken, moderator, []permissions.Check{admin}, nil, ""},
		{"cancelled context", cancelled, working, moderator, []permissions.Check{admin}, context.Canceled, ""},
		{"unknown permit", ctx, working, moderator, []permissions.Check{admin, {}}, nil, "unknown permit"},
		{
			name:     "group cycle",
			ctx:      ctx,
			checker:  working,
			actor:    stranger,
			checks:   []permissions.Check{admin, refModerate.On("cycle")},
			wantText: maxDepthError,
		},
		{
			name:     "depth error inside Except",
			ctx:      ctx,
			checker:  working,
			actor:    moderator,
			checks:   []permissions.Check{admin, refModerate.On("suspends")},
			wantText: maxDepthError,
		},
		{
			name:     "depth error through Via",
			ctx:      ctx,
			checker:  working,
			actor:    stranger,
			checks:   []permissions.Check{admin, refLogEdit.On("linked")},
			wantText: maxDepthError,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			each, eachErr := tc.checker.AllowedEach(signedIn(tc.ctx, tc.actor.String()), tc.checks...)
			forEach, forEachErr := tc.checker.AllowedForEach(tc.ctx, tc.checks[len(tc.checks)-1], users)

			for call, err := range map[string]error{"AllowedEach": eachErr, "AllowedForEach": forEachErr} {
				if errx.KindOf(err) != errx.Unavailable ||
					tc.cause != nil && !errors.Is(err, tc.cause) ||
					!strings.Contains(fmt.Sprint(err), tc.wantText) {
					t.Errorf("%s error = %v, want unavailable with %v %q", call, err, tc.cause, tc.wantText)
				}
			}
			if each != nil || forEach != nil {
				t.Errorf("results = %v and %v, want none", each, forEach)
			}
		})
	}

	t.Run("cancelled after the lookup", func(t *testing.T) {
		pool, statements := newLoggedPool(t, db)
		checker := permissions.NewChecker(NewRepository(pool), refModel)
		for call, check := range map[string]func(context.Context) error{
			"AllowedEach": func(ctx context.Context) error {
				_, err := checker.AllowedEach(signedIn(ctx, moderator.String()), admin)
				return err
			},
			"AllowedForEach": func(ctx context.Context) error {
				_, err := checker.AllowedForEach(ctx, admin, users)
				return err
			},
		} {
			lookupCtx, cancel := context.WithCancel(ctx)
			statements.onBatchEnd = cancel

			err := check(lookupCtx)

			if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, context.Canceled) {
				t.Errorf("%s error = %v, want unavailable wrapping context.Canceled", call, err)
			}
			cancel()
		}
	})
}

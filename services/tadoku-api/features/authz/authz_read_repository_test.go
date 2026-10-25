package authz

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestRead(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/read-0000000a")
	pool, statements := newLoggedPool(t, db)
	writer := permissions.NewRelationships(NewRepository(pool), testModel)
	owner, moderator, branchOwner, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contest, otherContest, team, otherTeam, logID := "c1", "c2", "g1", "g2", "l1"
	owned := ownersRelation.On(contest, owner)
	moderated := moderatorUsers.On(contest, moderator)
	contestTeam := moderatorGroups.On(contest, team)
	otherOwner := ownersRelation.On(otherContest, owner)
	link := contestsRelation.On(logID, contest)
	teamMember := memberUsers.On(team, owner)
	subteam := memberGroups.On(otherTeam, team)
	admin := siteAdminRelationship(owner)
	grantAll(t, ctx, writer, owned, moderated, contestTeam, otherOwner, link, teamMember, subteam, admin)
	grantAll(t, branchCtx, writer, ownersRelation.On(contest, branchOwner))

	cases := []struct {
		name   string
		filter permissions.Filter
		want   []permissions.Relationship
	}{
		{"on an object", permissions.OnObject(contests.Object(contest)), []permissions.Relationship{
			owned,
			moderated,
			contestTeam,
		}},
		{
			name:   "on an object, one relation",
			filter: permissions.OnObject(contests.Object(contest)).Relation(moderatorGroups),
			want:   []permissions.Relationship{moderated, contestTeam},
		},
		{"links on an object", permissions.OnObject(logs.Object(logID)), []permissions.Relationship{link}},
		{"for a user", permissions.ForUser(owner), []permissions.Relationship{
			owned,
			otherOwner,
			teamMember,
			admin,
		}},
		{
			name:   "for a user, in one namespace",
			filter: permissions.ForUser(owner).In(contests),
			want:   []permissions.Relationship{owned, otherOwner},
		},
		{
			name:   "for a user, one relation",
			filter: permissions.ForUser(owner).Relation(memberUsers),
			want:   []permissions.Relationship{teamMember},
		},
		{
			name:   "for a user, in a namespace and one relation",
			filter: permissions.ForUser(owner).In(groups).Relation(memberUsers),
			want:   []permissions.Relationship{teamMember},
		},
		{"only rows that store the user", permissions.ForUser(moderator), []permissions.Relationship{moderated}},
		{"for a user with no rows", permissions.ForUser(stranger), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			statements.take()

			got, err := writer.Read(ctx, tc.filter)

			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(rows(got), rows(tc.want)) {
				t.Errorf("read rows %s", rowDifference(rows(got), rows(tc.want)))
			}
			for _, relationship := range got {
				if !slices.Contains(tc.want, relationship) {
					t.Errorf("read %s, which is not equal to the relationship granted", row(relationship))
				}
			}
			if sent := statements.take(); len(sent) != 1 {
				t.Errorf("sent %d statements, want 1: %q", len(sent), sent)
			}
		})
	}

	t.Run("invalid filters", func(t *testing.T) {
		badges := permissions.MustDefine("badge", permissions.Relations{"holders": {permissions.Users}}, nil)
		badgeHolders := badges.MustUserRelation("holders")
		onContest := permissions.OnObject(contests.Object(contest))
		filters := map[string]permissions.Filter{
			"zero value":                       {},
			"object without a type or ID":      permissions.OnObject(permissions.Object{}),
			"object without an ID":             permissions.OnObject(contests.Object("")),
			"object of a namespace missing":    permissions.OnObject(badges.Object("b1")),
			"object narrowed to a namespace":   onContest.In(contests),
			"object with another relation":     onContest.Relation(memberUsers),
			"user in a namespace missing":      permissions.ForUser(owner).In(badges),
			"user in a namespace without it":   permissions.ForUser(owner).In(contests).Relation(memberUsers),
			"user with a missing relation":     permissions.ForUser(owner).Relation(badgeHolders),
			"user with a zero relation handle": permissions.ForUser(owner).Relation(permissions.UserRelation{}),
		}
		for name, filter := range filters {
			statements.take()

			_, err := writer.Read(ctx, filter)

			if err == nil {
				t.Errorf("%s: Read accepted an invalid filter", name)
			}
			if sent := statements.take(); len(sent) != 0 {
				t.Errorf("%s: sent %d statements, want none", name, len(sent))
			}
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		if _, err := writer.Read(t.Context(), permissions.ForUser(owner)); errx.KindOf(err) != errx.Unavailable {
			t.Errorf("Read without tenant error = %v, want unavailable", err)
		}
	})
}

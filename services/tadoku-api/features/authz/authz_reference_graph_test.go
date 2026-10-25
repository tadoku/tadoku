package authz

import (
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

var (
	refSiteAdmins       = refSites.MustUserRelation("admins")
	refMemberUsers      = refGroups.MustUserRelation("members")
	refMemberGroups     = refGroups.MustSetRelation("members")
	refParents          = refGroups.MustObjectRelation("parent")
	refOwners           = refContests.MustUserRelation("owners")
	refModeratorUsers   = refContests.MustUserRelation("moderators")
	refModeratorMembers = refContests.MustSetRelationOf("moderators", permissions.SetOf("group", "members"))
	refModeratorPermits = refContests.MustSetRelationOf("moderators", permissions.SetOf("group", "member"))
	refSuspendedUsers   = refContests.MustUserRelation("suspended")
	refSuspendedGroups  = refContests.MustSetRelation("suspended")
	refVerifiedUsers    = refContests.MustUserRelation("verified")
	refLogOwners        = refLogs.MustUserRelation("owner")
	refLogContests      = refLogs.MustObjectRelation("contest")

	refSiteAdmin = refSites.MustPermit("admin")
	refModerate  = refContests.MustPermit("moderate")
	refLogEdit   = refLogs.MustPermit("edit")
	refLogHide   = refLogs.MustPermit("hide")
)

// refQuestion is one permit on one object of the reference model, for the checker and the reference evaluator.
type refQuestion struct {
	object permissions.Object
	permit string
	check  permissions.Check
}

func refQuestions(pool graphPool) []refQuestion {
	var questions []refQuestion
	add := func(namespace permissions.Namespace, ids []string, permits ...string) {
		for _, id := range ids {
			for _, permit := range permits {
				questions = append(questions, refQuestion{
					object: namespace.Object(id),
					permit: permit,
					check:  namespace.MustPermit(permit).On(id),
				})
			}
		}
	}

	add(refSites, []string{siteRoot}, "admin")
	add(refGroups, pool.groups, "member")
	add(refContests, pool.contests, "edit", "moderate", "review")
	add(refLogs, pool.logs, "edit", "hide")
	return questions
}

// randomReferenceRelationship draws one relationship of the reference model. Group-to-group edges are rarer than
// the others, so most graphs stay within the depth limit and some do not.
func (p graphPool) randomReferenceRelationship(r *rand.Rand) permissions.Relationship {
	user := p.users[r.IntN(len(p.users))]
	group, otherGroup := p.groups[r.IntN(len(p.groups))], p.groups[r.IntN(len(p.groups))]
	contest, log := p.contests[r.IntN(len(p.contests))], p.logs[r.IntN(len(p.logs))]
	switch r.IntN(14) {
	case 0:
		return refSiteAdmins.On(siteRoot, user)
	case 1:
		return refMemberUsers.On(group, user)
	case 2:
		return refMemberGroups.On(group, otherGroup)
	case 3:
		return refParents.On(group, otherGroup)
	case 4:
		return refOwners.On(contest, user)
	case 5:
		return refModeratorUsers.On(contest, user)
	case 6:
		return refModeratorMembers.On(contest, group)
	case 7:
		return refModeratorPermits.On(contest, group)
	case 8:
		return refSuspendedUsers.On(contest, user)
	case 9:
		return refSuspendedGroups.On(contest, group)
	case 10:
		return refVerifiedUsers.On(contest, user)
	case 11, 12:
		return refLogContests.On(log, contest)
	default:
		return refLogOwners.On(log, user)
	}
}

func (p graphPool) randomReferenceGraph(r *rand.Rand, count int) []permissions.Relationship {
	graph := make([]permissions.Relationship, count)
	for i := range graph {
		graph[i] = p.randomReferenceRelationship(r)
	}
	return graph
}

func randomUsers(r *rand.Rand, users []uuid.UUID, count int) []uuid.UUID {
	picked := make([]uuid.UUID, count)
	for i := range picked {
		picked[i] = users[r.IntN(len(users))]
	}
	return picked
}

package authz

import (
	"fmt"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"pgregory.net/rapid"
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

func (q refQuestion) GoString() string {
	return fmt.Sprintf("%s:%s#%s", q.object.Type(), q.object.ID(), q.permit)
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

// referenceRelationship generates relationships of the reference model. Group-to-group edges are rarer than the
// others, so most graphs stay within the depth limit and some do not.
func (p graphPool) referenceRelationship() *rapid.Generator[permissions.Relationship] {
	user, contest := rapid.SampledFrom(p.users), rapid.SampledFrom(p.contests)
	group, log := rapid.SampledFrom(p.groups), rapid.SampledFrom(p.logs)
	return rapid.Custom(func(t *rapid.T) permissions.Relationship {
		switch rapid.IntRange(0, 13).Draw(t, "kind") {
		case 0:
			return refSiteAdmins.On(siteRoot, user.Draw(t, "user"))
		case 1:
			return refMemberUsers.On(group.Draw(t, "group"), user.Draw(t, "user"))
		case 2:
			return refMemberGroups.On(group.Draw(t, "group"), group.Draw(t, "subgroup"))
		case 3:
			return refParents.On(group.Draw(t, "group"), group.Draw(t, "parent"))
		case 4:
			return refOwners.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 5:
			return refModeratorUsers.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 6:
			return refModeratorMembers.On(contest.Draw(t, "contest"), group.Draw(t, "group"))
		case 7:
			return refModeratorPermits.On(contest.Draw(t, "contest"), group.Draw(t, "group"))
		case 8:
			return refSuspendedUsers.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 9:
			return refSuspendedGroups.On(contest.Draw(t, "contest"), group.Draw(t, "group"))
		case 10:
			return refVerifiedUsers.On(contest.Draw(t, "contest"), user.Draw(t, "user"))
		case 11, 12:
			return refLogContests.On(log.Draw(t, "log"), contest.Draw(t, "contest"))
		default:
			return refLogOwners.On(log.Draw(t, "log"), user.Draw(t, "user"))
		}
	})
}

package permissions

import (
	"testing"

	"github.com/google/uuid"
)

// TestRelationshipValidation gives Relationships no Tuples, so a relationship that reached storage would panic.
// Handles build only valid relationships for their own namespace, so the hand-built cases stand in for values a
// handle cannot produce.
func TestRelationshipValidation(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	writer := NewRelationships(nil, testModel)
	contestID, groupID, logID, user := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.New()
	contest, log := contests.Object(contestID), logs.Object(logID)
	members := subjectSet{object: groups.Object(groupID), relation: "members"}
	untyped, unnamed := Object{id: groupID}, subjectSet{object: contest}

	cases := []struct {
		name         string
		relationship Relationship
	}{
		{"namespace missing from the model", reviews.MustUserRelation("reviewers").On(contestID, user)},
		{"missing object ID", contests.MustUserRelation("owners").On("", user)},
		{"linked object without an ID", logs.MustObjectRelation("contests").On(logID, "")},
		{"subject set without an object ID", contests.MustSetRelation("moderators").On(contestID, "")},
		{"zero handle", UserRelation{}.On(contestID, user)},
		{"unknown relation", Relationship{object: contest, relation: "judges", subject: userSubject(user)}},
		{"group as owner", Relationship{object: contest, relation: "owners", subject: members}},
		{"user as linked object", Relationship{object: log, relation: "contests", subject: userSubject(user)}},
		{"linked object without a type", Relationship{object: contest, relation: "owners", subject: untyped}},
		{"subject set without a relation", Relationship{object: log, relation: "contests", subject: unnamed}},
		{"missing subject", Relationship{object: contest, relation: "owners"}},
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
		})
	}
}

func TestRelationshipsReadBack(t *testing.T) {
	t.Parallel()

	contestID, groupID, logID, userID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.New()
	for _, relationship := range []Relationship{
		contests.MustUserRelation("owners").On(contestID, userID),
		logs.MustObjectRelation("contests").On(logID, contestID),
		contests.MustSetRelation("moderators").On(contestID, groupID),
	} {
		read, err := relationshipOf(tupleOf(relationship))

		if err != nil || read != relationship {
			t.Errorf("read back %+v, %v; want %+v", read, err, relationship)
		}
	}

	tuple := tupleOf(contests.MustUserRelation("owners").On(contestID, userID))
	tuple.SubjectID = "not-a-uuid"
	if _, err := relationshipOf(tuple); err == nil {
		t.Error("read back a user subject that is not a UUID")
	}
}

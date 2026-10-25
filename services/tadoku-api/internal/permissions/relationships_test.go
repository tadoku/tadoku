package permissions

import (
	"testing"

	"github.com/google/uuid"
)

// TestRelationshipValidation gives Relationships no Tuples, so a relationship that reached storage would panic.
func TestRelationshipValidation(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	writer := NewRelationships(nil, testModel)
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
		})
	}
}

func TestRelationshipTuples(t *testing.T) {
	t.Parallel()

	contestID, groupID, logID, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	contest := contestObject(contestID)
	logObject := object{objectType: Log, id: logID.String()}

	cases := []struct {
		name         string
		relationship Relationship
		want         Tuple
	}{
		{
			name:         "user",
			relationship: Relationship{object: contest, relation: "owners", subject: userSubject(userID)},
			want: Tuple{
				ObjectType:  string(Contest),
				ObjectID:    contestID.String(),
				Relation:    "owners",
				SubjectType: UserSubjectType,
				SubjectID:   userID.String(),
			},
		},
		{
			name:         "subject set",
			relationship: Relationship{object: contest, relation: "moderators", subject: groupMembers(groupID)},
			want: Tuple{
				ObjectType:      string(Contest),
				ObjectID:        contestID.String(),
				Relation:        "moderators",
				SubjectType:     string(Group),
				SubjectID:       groupID.String(),
				SubjectRelation: "members",
			},
		},
		{
			name:         "linked object",
			relationship: Relationship{object: logObject, relation: "contests", subject: subjectSet{object: contest}},
			want: Tuple{
				ObjectType:  string(Log),
				ObjectID:    logID.String(),
				Relation:    "contests",
				SubjectType: string(Contest),
				SubjectID:   contestID.String(),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := testModel.validateRelationship(tc.relationship); err != nil {
				t.Fatalf("relationship is invalid: %v", err)
			}

			tuple := tupleOf(tc.relationship)

			if tuple != tc.want {
				t.Errorf("tuple = %+v, want %+v", tuple, tc.want)
			}
		})
	}
}

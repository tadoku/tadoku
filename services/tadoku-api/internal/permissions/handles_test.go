package permissions

import (
	"testing"

	"github.com/google/uuid"
)

var reviews = MustDefine(
	"review",
	Relations{
		"subjects":  {ObjectOf(Contest), ObjectOf(Log)},
		"reviewers": {Users, SetOf(Group, "members"), SetOf(Contest, "moderators")},
	},
	nil,
)

func TestHandlesRejectWrongNames(t *testing.T) {
	t.Parallel()

	groupMembers, groupOwners := SetOf(Group, "members"), SetOf(Group, "owners")

	cases := []struct {
		name   string
		handle func()
	}{
		{"unknown permit", func() { contests.MustPermit("judge") }},
		{"relation as permit", func() { contests.MustPermit("owners") }},
		{"unknown user relation", func() { contests.MustUserRelation("judges") }},
		{"permit as user relation", func() { contests.MustUserRelation("edit") }},
		{"user relation that holds objects", func() { logs.MustUserRelation("contests") }},
		{"unknown object relation", func() { logs.MustObjectRelation("parents") }},
		{"object relation that holds users", func() { contests.MustObjectRelation("owners") }},
		{"object relation with several object types", func() { reviews.MustObjectRelation("subjects") }},
		{"object relation of an undeclared type", func() { reviews.MustObjectRelationOf("subjects", ObjectOf(Group)) }},
		{"object relation of a subject set", func() { reviews.MustObjectRelationOf("reviewers", groupMembers) }},
		{"object relation of users", func() { reviews.MustObjectRelationOf("reviewers", Users) }},
		{"unknown set relation", func() { contests.MustSetRelation("judges") }},
		{"set relation that holds only users", func() { contests.MustSetRelation("owners") }},
		{"set relation with several subject sets", func() { reviews.MustSetRelation("reviewers") }},
		{"set relation of an undeclared set", func() { reviews.MustSetRelationOf("reviewers", groupOwners) }},
		{"set relation of an object type", func() { reviews.MustSetRelationOf("subjects", ObjectOf(Contest)) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("handle accepted a wrong name")
				}
			}()

			tc.handle()
		})
	}
}

func TestHandlesBuildChecksAndRelationships(t *testing.T) {
	t.Parallel()

	reviewID, contestID, groupID, logID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	userID := uuid.New()

	check := contests.MustPermit("edit").On(contestID)
	if check != (Check{object: contests.Object(contestID), permit: "edit"}) {
		t.Errorf("Permit.On = %+v, want edit on contest %s", check, contestID)
	}

	cases := []struct {
		name         string
		relationship Relationship
		want         Tuple
	}{
		{
			name:         "user",
			relationship: contests.MustUserRelation("moderators").On(contestID, userID),
			want: Tuple{
				ObjectType:  string(Contest),
				ObjectID:    contestID,
				Relation:    "moderators",
				SubjectType: UserSubjectType,
				SubjectID:   userID.String(),
			},
		},
		{
			name:         "linked object of the declared type",
			relationship: logs.MustObjectRelation("contests").On(logID, contestID),
			want: Tuple{
				ObjectType:  string(Log),
				ObjectID:    logID,
				Relation:    "contests",
				SubjectType: string(Contest),
				SubjectID:   contestID,
			},
		},
		{
			name:         "linked object of a named type",
			relationship: reviews.MustObjectRelationOf("subjects", ObjectOf(Log)).On(reviewID, logID),
			want: Tuple{
				ObjectType:  "review",
				ObjectID:    reviewID,
				Relation:    "subjects",
				SubjectType: string(Log),
				SubjectID:   logID,
			},
		},
		{
			name:         "subject set of the declared set",
			relationship: contests.MustSetRelation("moderators").On(contestID, groupID),
			want: Tuple{
				ObjectType:      string(Contest),
				ObjectID:        contestID,
				Relation:        "moderators",
				SubjectType:     string(Group),
				SubjectID:       groupID,
				SubjectRelation: "members",
			},
		},
		{
			name:         "subject set of a named set",
			relationship: reviews.MustSetRelationOf("reviewers", SetOf(Contest, "moderators")).On(reviewID, contestID),
			want: Tuple{
				ObjectType:      "review",
				ObjectID:        reviewID,
				Relation:        "reviewers",
				SubjectType:     string(Contest),
				SubjectID:       contestID,
				SubjectRelation: "moderators",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tuple := tupleOf(tc.relationship)

			if tuple != tc.want {
				t.Errorf("tuple = %+v, want %+v", tuple, tc.want)
			}
		})
	}
}

func TestRelationshipAccessors(t *testing.T) {
	t.Parallel()

	contestID, groupID, logID, userID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.New()

	cases := []struct {
		name         string
		relationship Relationship
		wantObject   Object
		wantRelation string
		wantUser     uuid.UUID
		wantTarget   Object
		wantSet      Object
		wantSetName  string
	}{
		{
			name:         "user",
			relationship: contests.MustUserRelation("owners").On(contestID, userID),
			wantObject:   contests.Object(contestID),
			wantRelation: "owners",
			wantUser:     userID,
		},
		{
			name:         "linked object",
			relationship: logs.MustObjectRelation("contests").On(logID, contestID),
			wantObject:   logs.Object(logID),
			wantRelation: "contests",
			wantTarget:   contests.Object(contestID),
		},
		{
			name:         "subject set",
			relationship: contests.MustSetRelation("moderators").On(contestID, groupID),
			wantObject:   contests.Object(contestID),
			wantRelation: "moderators",
			wantSet:      groups.Object(groupID),
			wantSetName:  "members",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			user, isUser := tc.relationship.User()
			target, isTarget := tc.relationship.Target()
			set, setRelation, isSet := tc.relationship.Set()

			if tc.relationship.Object() != tc.wantObject || tc.relationship.Relation() != tc.wantRelation {
				t.Errorf(
					"object and relation = %+v, %q; want %+v, %q",
					tc.relationship.Object(),
					tc.relationship.Relation(),
					tc.wantObject,
					tc.wantRelation,
				)
			}
			if user != tc.wantUser || isUser != (tc.wantUser != uuid.Nil) {
				t.Errorf("User() = %v, %t; want %v", user, isUser, tc.wantUser)
			}
			if target != tc.wantTarget || isTarget != (tc.wantTarget != Object{}) {
				t.Errorf("Target() = %+v, %t; want %+v", target, isTarget, tc.wantTarget)
			}
			if set != tc.wantSet || setRelation != tc.wantSetName || isSet != (tc.wantSetName != "") {
				t.Errorf("Set() = %+v, %q, %t; want %+v, %q", set, setRelation, isSet, tc.wantSet, tc.wantSetName)
			}
		})
	}

	object := contests.Object(contestID)
	if object.Type() != Contest || object.ID() != contestID {
		t.Errorf("Object = %s:%s, want %s:%s", object.Type(), object.ID(), Contest, contestID)
	}
}

package contests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreateContestParametersRejectsMissingSignedOwnerFields(t *testing.T) {
	now := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	valid := CreateContestParameters{
		ContestStart:            now,
		ContestEnd:              now,
		RegistrationEnd:         now,
		Title:                   "Valid title",
		ActivityTypeIDAllowList: []int32{1},
	}
	validOwnerID := uuid.MustParse("11111111-1111-4111-8111-111111111111")

	tests := []struct {
		name             string
		ownerID          uuid.UUID
		ownerDisplayName string
		message          string
	}{
		{name: "nil owner ID", ownerID: uuid.Nil, ownerDisplayName: "Reader One", message: "invalid contest OwnerUserID: must not be nil"},
		{name: "empty owner display name", ownerID: validOwnerID, message: "invalid contest OwnerUserDisplayName: must not be empty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := valid.validate(test.ownerID, test.ownerDisplayName, false, now)
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != test.message {
				t.Errorf("validate error=%v, want invalid input %q", err, test.message)
			}
		})
	}
}

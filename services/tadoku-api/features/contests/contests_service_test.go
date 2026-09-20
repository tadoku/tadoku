package contests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateParametersRejectsMissingSignedOwnerFields(t *testing.T) {
	now := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	valid := CreateParameters{
		ContestStart:            now,
		ContestEnd:              now,
		RegistrationEnd:         now,
		Title:                   "Valid title",
		ActivityTypeIDAllowList: []int32{1},
		ownerUserID:             uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ownerUserDisplayName:    "Reader One",
	}

	tests := []struct {
		name   string
		change func(*CreateParameters)
	}{
		{name: "nil owner ID", change: func(parameters *CreateParameters) { parameters.ownerUserID = uuid.Nil }},
		{name: "empty owner display name", change: func(parameters *CreateParameters) { parameters.ownerUserDisplayName = "" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			if err := parameters.validate(false, now); !errors.Is(err, ErrInvalidContest) {
				t.Errorf("validate error=%v, want invalid contest", err)
			}
		})
	}
}

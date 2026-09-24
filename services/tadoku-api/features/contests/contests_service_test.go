package contests

import (
	"errors"
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

func TestCheckContestCreationYearlyLimit(t *testing.T) {
	tests := []struct {
		name            string
		createdThisYear int64
		want            error
	}{
		{name: "below limit", createdThisYear: contestCreationYearlyLimit - 1, want: nil},
		{name: "at limit", createdThisYear: contestCreationYearlyLimit, want: ErrContestCreationForbidden},
		{name: "above limit", createdThisYear: contestCreationYearlyLimit + 1, want: ErrContestCreationForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkContestCreationYearlyLimit(test.createdThisYear)
			if !errors.Is(err, test.want) {
				t.Errorf("checkContestCreationYearlyLimit(%d) error=%v, want %v", test.createdThisYear, err, test.want)
			}
		})
	}
}

func TestCheckContestCreatorAccountAge(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	oneMonthAgo := now.AddDate(0, -1, 0)

	tests := []struct {
		name             string
		accountCreatedAt time.Time
		want             error
	}{
		{name: "older than one month", accountCreatedAt: oneMonthAgo.Add(-time.Second), want: nil},
		{name: "exactly one month", accountCreatedAt: oneMonthAgo, want: nil},
		{name: "younger than one month", accountCreatedAt: oneMonthAgo.Add(time.Second), want: ErrContestCreatorTooYoung},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkContestCreatorAccountAge(test.accountCreatedAt, now)
			if !errors.Is(err, test.want) {
				t.Errorf("checkContestCreatorAccountAge(%v, %v) error=%v, want %v", test.accountCreatedAt, now, err, test.want)
			}
		})
	}
}

func TestRegistrationUpsertParametersValidate(t *testing.T) {
	tests := []struct {
		name          string
		languageCodes []string
		want          string
	}{
		{name: "one language", languageCodes: []string{"jpa"}},
		{name: "three languages", languageCodes: []string{"jpa", "kor", "zho"}},
		{name: "no languages", languageCodes: nil, want: "invalid contest registration LanguageCodes: must contain at least one language"},
		{name: "four languages", languageCodes: []string{"jpa", "kor", "zho", "deu"}, want: "invalid contest registration LanguageCodes: must contain at most three languages"},
		{name: "duplicate language", languageCodes: []string{"jpa", "jpa"}, want: "invalid contest registration LanguageCodes: must not contain duplicates"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := RegistrationUpsertParameters{LanguageCodes: test.languageCodes}.Validate()

			if test.want == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
				return
			}
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != test.want {
				t.Errorf("Validate() error = %v, want invalid input %q", err, test.want)
			}
		})
	}
}

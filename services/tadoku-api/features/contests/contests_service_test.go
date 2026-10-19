package contests

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestApplyRegistrationDetachesUsingLanguageCodes(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	contestID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	registrationID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	contest := Contest{ID: contestID, Official: false}

	tests := []struct {
		name            string
		existingCodes   []string
		updatedCodes    []string
		wantEngLinks    int
		wantJpnLinks    int
		wantStoredCodes []string
	}{
		{
			name:            "drops unselected language codes without Languages hydration",
			existingCodes:   []string{"jpn", "eng"},
			updatedCodes:    []string{"jpn"},
			wantEngLinks:    0,
			wantJpnLinks:    1,
			wantStoredCodes: []string{"jpn"},
		},
		{
			name:            "does not detach when language codes are unchanged",
			existingCodes:   []string{"jpn", "eng"},
			updatedCodes:    []string{"jpn", "eng"},
			wantEngLinks:    1,
			wantJpnLinks:    1,
			wantStoredCodes: []string{"jpn", "eng"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tenantCtx := tenant.WithKey(t.Context(), tenant.Production())

			db, err := testpostgres.New(tenantCtx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})

			_, err = db.Pool.Exec(tenantCtx, `
				insert into users (tenant, id, display_name, created_at, updated_at)
				values
					('tadoku/prod', '11111111-1111-4111-8111-111111111111', 'Reader', '2026-01-01', '2026-01-01'),
					('tadoku/prod', '99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01');
				insert into contests (
					tenant, id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
					registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
				) values (
					'tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '99999999-9999-4999-8999-999999999999', 'Owner', true,
					'2026-09-01', '2026-09-12', '2026-08-31', 'Registration fixture', '{1}', false,
					'2026-01-01', '2026-01-01'
				);
				insert into contest_registrations (
					tenant, id, contest_id, user_id, language_codes, created_at, updated_at
				)
				values (
					'tadoku/prod', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
					'11111111-1111-4111-8111-111111111111', '{jpn,eng}', '2026-01-01', '2026-01-01'
				);
				insert into logs (
					tenant, id, user_id, language_code, log_activity_id, duration_seconds, computed_score,
					eligible_official_leaderboard, created_at, updated_at
				) values
					('tadoku/prod', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1',
					 '11111111-1111-4111-8111-111111111111', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
					('tadoku/prod', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2',
					 '11111111-1111-4111-8111-111111111111', 'jpn', 1, 60, 1, true, '2026-09-01', '2026-09-01');
				insert into contest_logs (tenant, contest_id, log_id, duration_seconds, computed_score)
				values
					('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', 60, 1),
					('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2', 60, 1);
			`)
			if err != nil {
				t.Fatal(err)
			}

			service := NewService(NewContestsRepository(db.Pool))
			existing := &Registration{
				ID:            registrationID,
				ContestID:     contestID,
				UserID:        userID,
				LanguageCodes: test.existingCodes,
			}
			updated := Registration{
				ID:            registrationID,
				ContestID:     contestID,
				UserID:        userID,
				LanguageCodes: test.updatedCodes,
				CreatedAt:     createdAt,
				UpdatedAt:     now,
			}

			if _, err := service.ApplyRegistration(tenantCtx, updated, existing, contest); err != nil {
				t.Fatal(err)
			}

			var engLinks, jpnLinks int
			if err := db.Pool.QueryRow(tenantCtx, `
				select
					(select count(*) from contest_logs cl
					 join logs l on l.id = cl.log_id
					 where l.language_code = 'eng'),
					(select count(*) from contest_logs cl
					 join logs l on l.id = cl.log_id
					 where l.language_code = 'jpn')
			`).Scan(&engLinks, &jpnLinks); err != nil {
				t.Fatal(err)
			}
			if engLinks != test.wantEngLinks || jpnLinks != test.wantJpnLinks {
				t.Errorf("contest log links eng=%d jpn=%d, want eng=%d jpn=%d",
					engLinks, jpnLinks, test.wantEngLinks, test.wantJpnLinks)
			}

			stored, err := service.contests.FindRegistrationForUser(tenantCtx, userID, contestID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(stored.LanguageCodes, test.wantStoredCodes) {
				t.Errorf("stored language codes=%v, want %v", stored.LanguageCodes, test.wantStoredCodes)
			}
		})
	}
}

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

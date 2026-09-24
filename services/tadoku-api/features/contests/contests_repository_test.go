package contests

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestContestsRepositoryDiscoveryQueries(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(t.Context(), `
		insert into users (id, display_name, created_at, updated_at, deleted_at)
		values
			('11111111-1111-4111-8111-111111111111', 'Owner One', '2026-01-01', '2026-01-01', null),
			('22222222-2222-4222-8222-222222222222', 'Owner Two', '2026-01-01', '2026-01-01', '2026-09-01');

		insert into contests (
			id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, "description", language_code_allow_list,
			activity_type_id_allow_list, official, created_at, updated_at, deleted_at
		)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'stale', false,
			 '2026-01-01', '2026-01-31', '2026-01-15', 'Public official', null, null, '{1}', true,
			 '2026-01-01', '2026-01-01', null),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '22222222-2222-4222-8222-222222222222', 'stale', true,
			 '2026-02-01', '2026-02-28', '2026-02-15', 'Private official', 'private', '{jpn,eng}', '{2,1}', true,
			 '2026-02-01', '2026-02-01', null),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '11111111-1111-4111-8111-111111111111', 'stale', true,
			 '2026-03-01', '2026-03-31', '2026-03-15', 'Private unofficial', null, '{}', '{3}', false,
			 '2026-03-01', '2026-03-01', null),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', '22222222-2222-4222-8222-222222222222', 'stale', false,
			 '2027-01-01', '2027-01-31', '2026-12-15', 'Deleted future official', null, '{eng}', '{5}', true,
			 '2026-04-01', '2026-04-01', '2026-04-02'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', '33333333-3333-4333-8333-333333333333', 'missing', false,
			 '2026-02-15', '2026-02-28', '2026-02-01', 'Orphan official', null, null, '{1}', true,
			 '2026-02-15', '2026-02-15', null)`)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewContestsRepository(db.Pool)
	parameters := ListParameters{
		Official: true,
		PageSize: 10,
	}
	items, total, err := repository.ListContests(t.Context(), parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "Public official" || total != 3 {
		t.Errorf("public list=%+v total=%d, want one visible of three matching official contests", items, total)
	}

	parameters.Page = 1
	items, total, err = repository.ListContests(t.Context(), parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 || total != 3 {
		t.Errorf("out-of-range list=%+v total=%d, want empty page with three matches", items, total)
	}

	items, total, err = repository.ListContests(t.Context(), ListParameters{
		Official: false,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 || total != 1 {
		t.Errorf("hidden-only list=%+v total=%d, want empty page with one private match", items, total)
	}

	orphanOwnerID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	items, total, err = repository.ListContests(t.Context(), ListParameters{
		UserID:   &orphanOwnerID,
		Official: true,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 || total != 1 {
		t.Errorf("orphan-only list=%+v total=%d, want empty page with one orphan match", items, total)
	}

	ownerID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	parameters = ListParameters{
		UserID:   &ownerID,
		Official: false,
		PageSize: 10,
	}
	items, total, err = repository.ListContests(t.Context(), parameters)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "Private unofficial" || total != 1 {
		t.Errorf("owner list=%+v total=%d, want private owner contest", items, total)
	}

	missingOwnerID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	items, total, err = repository.ListContests(t.Context(), ListParameters{
		UserID:   &missingOwnerID,
		Official: false,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 || total != 0 {
		t.Errorf("zero-match list=%+v total=%d, want empty page and zero total", items, total)
	}

	deletedID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4")
	if _, err := repository.FindContestByID(t.Context(), FindParameters{ID: deletedID}); !errors.Is(err, ErrContestNotFound) {
		t.Errorf("deleted contest error=%v, want not found", err)
	}
	deleted, err := repository.FindContestByID(t.Context(), FindParameters{ID: deletedID, includeDeleted: true})
	if err != nil {
		t.Fatal(err)
	}
	if !deleted.Deleted || deleted.OwnerUserDisplayName != "Deleted organizer" {
		t.Errorf("deleted contest=%+v", deleted)
	}

	privateID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	private, err := repository.FindContestByID(t.Context(), FindParameters{ID: privateID})
	if err != nil {
		t.Fatal(err)
	}
	languages, err := repository.ListLanguagesForContest(t.Context(), private.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(languages) != 2 {
		t.Fatalf("languages=%+v, want two", languages)
	}
	if got := []string{languages[0].Name, languages[1].Name}; !reflect.DeepEqual(got, []string{"English", "Japanese"}) {
		t.Errorf("language names=%v", got)
	}

	latest, err := repository.FindLatestOfficialContest(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if latest.ID != deletedID {
		t.Errorf("latest official=%s, want future deleted contest %s", latest.ID, deletedID)
	}
}

func TestContestsRepositoryCreationTransaction(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	repository := NewContestsRepository(db.Pool)
	now := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	description := "Created through the repository"
	contest := Contest{
		ID:                      uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		ContestStart:            time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
		ContestEnd:              time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC),
		RegistrationEnd:         time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
		Title:                   "October contest",
		Description:             &description,
		OwnerUserID:             uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		OwnerUserDisplayName:    "Reader One",
		LanguageCodeAllowList:   []string{"jpn"},
		ActivityTypeIDAllowList: []int32{1, 2},
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if _, err := db.Pool.Exec(t.Context(), `
		insert into users (id, display_name, created_at, updated_at)
		values ($1, $2, $3, $3)`, contest.OwnerUserID, contest.OwnerUserDisplayName, now); err != nil {
		t.Fatal(err)
	}

	var created *Contest
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.CreateContest(ctx, contest); err != nil {
			return err
		}
		var err error
		created, err = repository.FindContestByID(ctx, FindParameters{ID: contest.ID})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != contest.ID || created.OwnerUserDisplayName != "Reader One" || !reflect.DeepEqual(created.LanguageCodeAllowList, []string{"jpn"}) {
		t.Errorf("created contest=%+v", created)
	}

	rolledBack := contest
	rolledBack.ID = uuid.MustParse("88888888-8888-4888-8888-888888888888")
	rollbackErr := errors.New("force rollback")
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.CreateContest(ctx, rolledBack); err != nil {
			return err
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("rollback error=%v, want %v", err, rollbackErr)
	}
	if _, err := repository.FindContestByID(t.Context(), FindParameters{ID: rolledBack.ID}); !errors.Is(err, ErrContestNotFound) {
		t.Errorf("rolled-back contest error=%v, want contest not found", err)
	}
}

func TestContestsRepositoryCountsEveryContestCreatedByUserInYear(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(t.Context(), `
		insert into contests (
			id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at, deleted_at
		)
		select
			('10000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
			'11111111-1111-4111-8111-111111111111', 'Owner One', n % 2 = 0,
			'2025-01-01', '2025-01-31', '2025-01-01', 'Current year ' || n, '{}', n % 2 = 1,
			'2026-01-01'::timestamp + n * interval '1 day', '2026-01-01',
			case when n = 12 then '2026-02-01'::timestamp end
		from generate_series(1, 12) as n;

		insert into contests (
			id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
		)
		values
			('20000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Owner One', false,
			 '2026-01-01', '2026-01-31', '2026-01-01', 'Prior year', '{}', true, '2025-12-31', '2025-12-31'),
			('20000000-0000-4000-8000-000000000002', '22222222-2222-4222-8222-222222222222', 'Owner Two', false,
			 '2026-01-01', '2026-01-31', '2026-01-01', 'Other owner', '{}', true, '2026-01-01', '2026-01-01')`)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewContestsRepository(db.Pool)
	ownerID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	count, err := repository.CountContestsCreatedByUserForYear(t.Context(), ownerID, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if count != 12 {
		t.Errorf("2026 count = %d, want 12 including private, unofficial, and deleted contests", count)
	}

	count, err = repository.CountContestsCreatedByUserForYear(t.Context(), ownerID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("2025 count = %d, want 1", count)
	}
}

func TestContestsRepositoryRegistrationPersistenceAndTransaction(t *testing.T) {
	t.Parallel()

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(t.Context(), `
		insert into users (id, display_name, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'Reader', '2026-01-01', '2026-01-01'),
			('99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01');
		insert into contests (
			id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
		) values (
			'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '99999999-9999-4999-8999-999999999999', 'stale', true,
			'2026-09-01', '2026-09-12', '2026-08-31', 'Registration fixture', '{2,1}', true,
			'2026-01-01', '2026-01-01'
		);
		insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at)
		values (
			'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
			'11111111-1111-4111-8111-111111111111', '{jpn,eng}', '2026-01-01', '2026-01-01'
		);
		insert into logs (
			id, user_id, language_code, log_activity_id, duration_seconds, computed_score,
			eligible_official_leaderboard, created_at, updated_at
		) values
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 60, 1, true, '2026-09-01', '2026-09-01');
		insert into contest_logs (contest_id, log_id, duration_seconds, computed_score)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', 60, 1),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2', 60, 1);
	`)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewContestsRepository(db.Pool)
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	contestID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	registration, err := repository.FindRegistrationForUser(t.Context(), userID, contestID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(registration.LanguageCodes, []string{"jpn", "eng"}) {
		t.Errorf("registration language codes=%v", registration.LanguageCodes)
	}
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !registration.CreatedAt.Equal(createdAt) || !registration.UpdatedAt.Equal(createdAt) {
		t.Errorf("registration timestamps=%s/%s, want %s", registration.CreatedAt, registration.UpdatedAt, createdAt)
	}
	if registration.Contest != nil {
		t.Errorf("registration contest=%+v, want nil", registration.Contest)
	}

	withContest, err := repository.FindRegistrationWithContestForUser(t.Context(), userID, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if withContest.ID != registration.ID || withContest.Contest == nil ||
		withContest.Contest.Title != "Registration fixture" || !withContest.Contest.Official {
		t.Errorf("registration with contest=%+v", withContest)
	}

	ongoing, err := repository.ListOngoingRegistrations(t.Context(), userID, time.Date(2026, 9, 12, 23, 59, 59, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(ongoing) != 1 || !reflect.DeepEqual(ongoing[0].LanguageCodes, []string{"jpn", "eng"}) ||
		!reflect.DeepEqual(ongoing[0].Contest.allowedActivityIDs, []int32{2, 1}) {
		t.Errorf("ongoing registration=%+v", ongoing)
	}

	ongoing, err = repository.ListOngoingRegistrations(t.Context(), userID, time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(ongoing) != 0 {
		t.Errorf("registration remains ongoing after end day: %+v", ongoing)
	}

	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	updatedRegistration := *registration
	updatedRegistration.LanguageCodes = []string{"jpn"}
	updatedRegistration.UpdatedAt = now
	removedLanguages := []string{"eng"}
	insertRefreshes := func(ctx context.Context) error {
		if err := repository.InsertContestScoreRefresh(ctx, userID, contestID); err != nil {
			return err
		}
		return repository.InsertOfficialScoresRefresh(ctx, userID, 2026)
	}

	rollbackErr := errors.New("force registration rollback")
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.DetachContestLogsForLanguages(ctx, userID, contestID, removedLanguages); err != nil {
			return err
		}
		if err := insertRefreshes(ctx); err != nil {
			return err
		}
		if err := repository.UpsertRegistration(ctx, updatedRegistration); err != nil {
			return err
		}
		if err := insertRefreshes(ctx); err != nil {
			return err
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("rollback error=%v", err)
	}

	var links, events int
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from contest_logs`).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from leaderboard_outbox`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if links != 2 || events != 0 {
		t.Errorf("after rollback links=%d events=%d, want 2 and 0", links, events)
	}

	registration, err = repository.FindRegistrationForUser(t.Context(), userID, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(registration.LanguageCodes, []string{"jpn", "eng"}) {
		t.Errorf("rolled-back language codes=%+v", registration.LanguageCodes)
	}

	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.DetachContestLogsForLanguages(ctx, userID, contestID, removedLanguages); err != nil {
			return err
		}
		if err := insertRefreshes(ctx); err != nil {
			return err
		}
		if err := repository.UpsertRegistration(ctx, updatedRegistration); err != nil {
			return err
		}
		return insertRefreshes(ctx)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Pool.QueryRow(t.Context(), `select count(*) from contest_logs`).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select count(*) from leaderboard_outbox`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if links != 1 || events != 4 {
		t.Errorf("after update links=%d events=%d, want 1 and 4", links, events)
	}

	registration, err = repository.FindRegistrationForUser(t.Context(), userID, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(registration.LanguageCodes, []string{"jpn"}) {
		t.Errorf("updated language codes=%+v", registration.LanguageCodes)
	}
}

package contests

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
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

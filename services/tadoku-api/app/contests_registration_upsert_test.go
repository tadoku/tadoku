package app

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestUpsertContestRegistrationConcurrentFirstTime(t *testing.T) {
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

	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	ownerID := uuid.MustParse("99999999-9999-4999-8999-999999999999")
	contestID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

	_, err = db.Pool.Exec(tenantCtx, `
		insert into users (tenant, id, display_name, created_at, updated_at)
		values ('tadoku/prod', $1, 'Owner', '2026-01-01', '2026-01-01');
		insert into contests (
			tenant, id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
		) values (
			'tadoku/prod', $2, $1, 'Owner', false,
			'2026-09-01', '2026-09-30', '2026-09-30', 'Concurrent registration', '{1}', false,
			'2026-01-01', '2026-01-01'
		);
	`, ownerID, contestID)
	if err != nil {
		t.Fatal(err)
	}

	application := New(Dependencies{
		JobQueue:    jobqueue.NewService(jobqueue.NewRepository(db.AppPool)),
		Contests:    contests.NewService(contests.NewContestsRepository(db.AppPool)),
		Languages:   languages.NewService(languages.NewLanguagesRepository(db.AppPool)),
		Profile:     profile.NewService(profile.NewRepository(db.AppPool), nil, nil),
		DB:          db.AppPool,
		Permissions: permissions.NewKetoChecker(nil),
	})

	ctx := identity.WithUser(tenantCtx, &identity.User{
		Subject:     userID.String(),
		DisplayName: "Reader",
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	parameters := ContestRegistrationUpsertParameters{
		ContestID:     contestID,
		LanguageCodes: []string{"jpn"},
	}

	const workers = 8
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- application.UpsertContestRegistration(ctx, parameters)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("UpsertContestRegistration error=%v, want nil", err)
		}
	}

	var count int
	if err := db.Pool.QueryRow(tenantCtx, `
		select count(*) from contest_registrations
		where tenant = 'tadoku/prod' and user_id = $1 and contest_id = $2 and deleted_at is null
	`, userID, contestID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("contest_registrations rows=%d, want 1", count)
	}
}

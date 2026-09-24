package logs

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/leaderboardoutbox"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

var (
	testUserID      = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testOtherUserID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
)

func newTestLogsRepository(t *testing.T) (*LogsRepository, *testpostgres.Database) {
	t.Helper()

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
		insert into users (id, display_name)
		values
			('11111111-1111-4111-8111-111111111111', 'Reader'),
			('22222222-2222-4222-8222-222222222222', 'Other')`)
	if err != nil {
		t.Fatal(err)
	}

	return NewLogsRepository(db.Pool), db
}

func createDurationLog(t *testing.T, repository *LogsRepository, id, userID uuid.UUID, createdAt time.Time) {
	t.Helper()

	duration := int32(600)
	err := repository.CreateLog(t.Context(), logMutation{
		ID:           id,
		UserID:       userID,
		LanguageCode: "jpn",
		ActivityID:   2,
		Tracking: Tracking{
			DurationSeconds: &duration,
			Score:           4,
		},
		Year: int16(createdAt.Year()),
		Now:  createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogsRepositoryFindLogMapsRows(t *testing.T) {
	t.Parallel()
	repository, db := newTestLogsRepository(t)

	units, err := repository.ListUnits(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var unit *Unit
	for i := range units {
		if units[i].LogActivityID == 1 && units[i].LanguageCode == nil {
			unit = &units[i]
			break
		}
	}
	if unit == nil {
		t.Fatal("no seeded reading unit without language")
	}

	var ruleSetID uuid.UUID
	if err := db.Pool.QueryRow(t.Context(), `select id from scoring_rule_sets limit 1`).Scan(&ruleSetID); err != nil {
		t.Fatal(err)
	}

	amountLogID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	durationLogID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	createdAt := time.Date(2026, 9, 1, 12, 30, 0, 0, time.UTC)
	description := "Chapter 3"
	amount := float32(12.5)
	modifier := unit.Modifier
	ruleID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1")
	tracking := Tracking{
		UnitID:    &unit.ID,
		UnitKey:   unit.Key,
		Amount:    &amount,
		Modifier:  &modifier,
		Score:     20,
		RuleSetID: &ruleSetID,
		RuleIDs:   []uuid.UUID{ruleID},
		Rates:     []float32{1.6},
		Source:    "amount",
	}

	err = repository.CreateLog(t.Context(), logMutation{
		ID:           amountLogID,
		UserID:       testUserID,
		LanguageCode: "jpn",
		ActivityID:   1,
		Description:  &description,
		Tracking:     tracking,
		Year:         2026,
		Now:          createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{"book", "fiction"} {
		if err := repository.InsertTag(t.Context(), amountLogID, testUserID, tag); err != nil {
			t.Fatal(err)
		}
	}
	createDurationLog(t, repository, durationLogID, testUserID, createdAt)

	found, err := repository.FindLog(t.Context(), amountLogID, false)
	if err != nil {
		t.Fatal(err)
	}
	displayName := "Reader"
	want := &Log{
		ID:              amountLogID,
		UserID:          testUserID,
		UserDisplayName: &displayName,
		Description:     &description,
		LanguageCode:    "jpn",
		LanguageName:    "Japanese",
		Activity:        activities.Activity{ID: 1},
		UnitID:          unit.ID,
		UnitKey:         unit.Key,
		UnitName:        unit.Name,
		Tags:            []string{"book", "fiction"},
		Amount:          amount,
		Modifier:        modifier,
		Score:           20,
		CreatedAt:       found.CreatedAt,
		Tracking:        tracking,
	}
	if !found.CreatedAt.Equal(createdAt) {
		t.Errorf("created_at=%s, want %s", found.CreatedAt, createdAt)
	}
	if !reflect.DeepEqual(found, want) {
		t.Errorf("amount log=%+v, want %+v", found, want)
	}

	found, err = repository.FindLog(t.Context(), durationLogID, false)
	if err != nil {
		t.Fatal(err)
	}
	duration := int32(600)
	want = &Log{
		ID:              durationLogID,
		UserID:          testUserID,
		UserDisplayName: &displayName,
		LanguageCode:    "jpn",
		LanguageName:    "Japanese",
		Activity:        activities.Activity{ID: 2},
		Tags:            []string{},
		Score:           4,
		DurationSeconds: &duration,
		CreatedAt:       found.CreatedAt,
		Tracking: Tracking{
			DurationSeconds: &duration,
			Score:           4,
			RuleIDs:         []uuid.UUID{},
		},
	}
	if !reflect.DeepEqual(found, want) {
		t.Errorf("duration log=%+v, want %+v", found, want)
	}

	if _, err := repository.FindLog(t.Context(), uuid.New(), true); !errors.Is(err, ErrLogNotFound) {
		t.Errorf("missing log error=%v, want not found", err)
	}
}

func TestLogsRepositorySoftDeleteVisibility(t *testing.T) {
	t.Parallel()
	repository, db := newTestLogsRepository(t)

	logID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	frozenLogID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	deletedAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	createDurationLog(t, repository, logID, testUserID, createdAt)
	createDurationLog(t, repository, frozenLogID, testUserID, createdAt)

	if _, err := db.Pool.Exec(t.Context(), `update logs set frozen_at = '2026-09-01 13:00' where id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2'`); err != nil {
		t.Fatal(err)
	}

	if err := repository.LockLog(t.Context(), logID); err != nil {
		t.Errorf("lock live log error=%v, want nil", err)
	}
	if err := repository.LockLog(t.Context(), frozenLogID); !errors.Is(err, ErrLogFrozen) {
		t.Errorf("lock frozen log error=%v, want frozen", err)
	}

	for _, id := range []uuid.UUID{logID, frozenLogID} {
		if err := repository.SoftDelete(t.Context(), id, deletedAt); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := repository.FindLog(t.Context(), logID, false); !errors.Is(err, ErrLogNotFound) {
		t.Errorf("deleted log without deleted rows error=%v, want not found", err)
	}
	if err := repository.LockLog(t.Context(), logID); !errors.Is(err, ErrLogNotFound) {
		t.Errorf("lock deleted log error=%v, want not found", err)
	}
	found, err := repository.FindLog(t.Context(), logID, true)
	if err != nil {
		t.Fatal(err)
	}
	if !found.Deleted {
		t.Errorf("deleted log=%+v, want deleted", found)
	}

	frozen, err := repository.FindLog(t.Context(), frozenLogID, false)
	if err != nil {
		t.Fatal(err)
	}
	if frozen.Deleted {
		t.Errorf("frozen log=%+v, want soft delete ignored", frozen)
	}
}

func TestLogsRepositoryListUserLogsPaging(t *testing.T) {
	t.Parallel()
	repository, _ := newTestLogsRepository(t)

	oldest := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	middle := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	newest := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3")
	deleted := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4")
	otherUser := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5")
	day := func(d int) time.Time { return time.Date(2026, 9, d, 12, 0, 0, 0, time.UTC) }

	createDurationLog(t, repository, oldest, testUserID, day(1))
	createDurationLog(t, repository, middle, testUserID, day(2))
	createDurationLog(t, repository, newest, testUserID, day(3))
	createDurationLog(t, repository, deleted, testUserID, day(4))
	createDurationLog(t, repository, otherUser, testOtherUserID, day(5))
	if err := repository.SoftDelete(t.Context(), deleted, day(6)); err != nil {
		t.Fatal(err)
	}

	userID := testUserID
	tests := []struct {
		name           string
		page           int
		includeDeleted bool
		wantIDs        []uuid.UUID
		wantTotal      int
		wantNext       string
	}{
		{name: "first page", page: 0, wantIDs: []uuid.UUID{newest, middle}, wantTotal: 3, wantNext: "1"},
		{name: "last page", page: 1, wantIDs: []uuid.UUID{oldest}, wantTotal: 3},
		{name: "past end", page: 2, wantIDs: nil},
		{name: "with deleted", page: 0, includeDeleted: true, wantIDs: []uuid.UUID{deleted, newest}, wantTotal: 4, wantNext: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := repository.ListUserLogs(t.Context(), ListParameters{
				UserID:         &userID,
				IncludeDeleted: tt.includeDeleted,
				PageSize:       2,
				Page:           tt.page,
			})
			if err != nil {
				t.Fatal(err)
			}

			var ids []uuid.UUID
			for _, log := range list.Logs {
				ids = append(ids, log.ID)
				if log.UserID != testUserID {
					t.Errorf("log %s user=%s, want %s", log.ID, log.UserID, testUserID)
				}
				if log.Deleted != (log.ID == deleted) {
					t.Errorf("log %s deleted=%t", log.ID, log.Deleted)
				}
			}
			if !reflect.DeepEqual(ids, tt.wantIDs) {
				t.Errorf("ids=%v, want %v", ids, tt.wantIDs)
			}
			if list.TotalSize != tt.wantTotal || list.NextPageToken != tt.wantNext {
				t.Errorf("total=%d next=%q, want total=%d next=%q", list.TotalSize, list.NextPageToken, tt.wantTotal, tt.wantNext)
			}
		})
	}
}

func TestLogsRepositoryContestLogs(t *testing.T) {
	t.Parallel()
	repository, db := newTestLogsRepository(t)

	_, err := db.Pool.Exec(t.Context(), `
		insert into contests (
			id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
		)
		values
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '99999999-9999-4999-8999-999999999999', 'Owner', false,
			 '2026-09-01', '2026-09-30', '2026-09-15', 'Listed', '{2}', false, '2026-08-01', '2026-08-01'),
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '99999999-9999-4999-8999-999999999999', 'Owner', false,
			 '2026-09-01', '2026-09-30', '2026-09-15', 'Other', '{2}', false, '2026-08-01', '2026-08-01');
		insert into contest_registrations (id, contest_id, user_id, language_codes)
		values
			('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1',
			 '11111111-1111-4111-8111-111111111111', '{jpn}'),
			('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1',
			 '22222222-2222-4222-8222-222222222222', '{jpn}'),
			('eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2',
			 '11111111-1111-4111-8111-111111111111', '{jpn}')`)
	if err != nil {
		t.Fatal(err)
	}

	contestID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	registration := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee1")
	otherRegistration := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee2")
	otherContestRegistration := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeee3")
	oldest := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	middle := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2")
	newest := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3")
	deleted := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4")
	otherUser := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5")
	otherContest := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa6")
	unattached := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa7")
	day := func(d int) time.Time { return time.Date(2026, 9, d, 12, 0, 0, 0, time.UTC) }

	duration := int32(600)
	attach := func(logID, registrationID uuid.UUID) {
		t.Helper()

		err := repository.CreateContestLog(t.Context(), logID, ContestTracking{
			RegistrationID: registrationID,
			Tracking: Tracking{
				DurationSeconds: &duration,
				Score:           7,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, log := range []struct {
		id           uuid.UUID
		userID       uuid.UUID
		registration uuid.UUID
		day          int
	}{
		{id: oldest, userID: testUserID, registration: registration, day: 1},
		{id: middle, userID: testUserID, registration: registration, day: 2},
		{id: newest, userID: testUserID, registration: registration, day: 3},
		{id: deleted, userID: testUserID, registration: registration, day: 4},
		{id: otherUser, userID: testOtherUserID, registration: otherRegistration, day: 5},
		{id: otherContest, userID: testUserID, registration: otherContestRegistration, day: 6},
	} {
		createDurationLog(t, repository, log.id, log.userID, day(log.day))
		attach(log.id, log.registration)
	}
	createDurationLog(t, repository, unattached, testUserID, day(7))
	if err := repository.SoftDelete(t.Context(), deleted, day(8)); err != nil {
		t.Fatal(err)
	}

	userID := testUserID
	tests := []struct {
		name           string
		userID         *uuid.UUID
		page           int
		includeDeleted bool
		wantIDs        []uuid.UUID
		wantTotal      int
		wantNext       string
	}{
		{name: "all users first page", page: 0, wantIDs: []uuid.UUID{otherUser, newest}, wantTotal: 4, wantNext: "1"},
		{name: "all users last page", page: 1, wantIDs: []uuid.UUID{middle, oldest}, wantTotal: 4},
		{name: "past end", page: 2, wantIDs: nil},
		{name: "one user", userID: &userID, page: 0, wantIDs: []uuid.UUID{newest, middle}, wantTotal: 3, wantNext: "1"},
		{name: "one user with deleted", userID: &userID, page: 0, includeDeleted: true, wantIDs: []uuid.UUID{deleted, newest}, wantTotal: 4, wantNext: "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list, err := repository.ListContestLogs(t.Context(), ListParameters{
				UserID:         tt.userID,
				ContestID:      contestID,
				IncludeDeleted: tt.includeDeleted,
				PageSize:       2,
				Page:           tt.page,
			})
			if err != nil {
				t.Fatal(err)
			}

			var ids []uuid.UUID
			for _, log := range list.Logs {
				ids = append(ids, log.ID)
				if log.Deleted != (log.ID == deleted) {
					t.Errorf("log %s deleted=%t", log.ID, log.Deleted)
				}
				if log.Score != 7 || log.DurationSeconds == nil || *log.DurationSeconds != duration {
					t.Errorf("log %s score=%v duration=%v, want contest tracking", log.ID, log.Score, log.DurationSeconds)
				}
				wantName := "Reader"
				if log.UserID == testOtherUserID {
					wantName = "Other"
				}
				if log.UserDisplayName == nil || *log.UserDisplayName != wantName {
					t.Errorf("log %s display name=%v, want %q", log.ID, log.UserDisplayName, wantName)
				}
			}
			if !reflect.DeepEqual(ids, tt.wantIDs) {
				t.Errorf("ids=%v, want %v", ids, tt.wantIDs)
			}
			if list.TotalSize != tt.wantTotal || list.NextPageToken != tt.wantNext {
				t.Errorf("total=%d next=%q, want total=%d next=%q", list.TotalSize, list.NextPageToken, tt.wantTotal, tt.wantNext)
			}
		})
	}

	contestEnd := time.Date(2026, 9, 30, 23, 0, 0, 0, time.UTC)
	afterContestEnd := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		id   uuid.UUID
		now  time.Time
		want bool
	}{
		{id: oldest, now: contestEnd, want: true},
		{id: oldest, now: afterContestEnd, want: false},
		{id: deleted, now: afterContestEnd, want: false},
		{id: unattached, now: afterContestEnd, want: true},
	} {
		got, err := repository.CanDelete(t.Context(), tt.id, tt.now)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("can delete %s at %s=%t, want %t", tt.id, tt.now, got, tt.want)
		}
	}
}

func TestLogsRepositoryInsertOutbox(t *testing.T) {
	t.Parallel()
	repository, db := newTestLogsRepository(t)

	contestID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	year := int16(2026)
	if err := repository.InsertOutbox(t.Context(), testUserID, nil, nil, leaderboardoutbox.RefreshOfficialScores); err != nil {
		t.Fatal(err)
	}
	if err := repository.InsertOutbox(t.Context(), testOtherUserID, &contestID, &year, leaderboardoutbox.RefreshContestScore); err != nil {
		t.Fatal(err)
	}

	type outboxRow struct {
		EventType   string
		UserID      uuid.UUID
		ContestID   *uuid.UUID
		Year        *int16
		Unprocessed bool
	}
	rows, err := db.Pool.Query(t.Context(), `
		select event_type, user_id, contest_id, year, processed_at is null
		from leaderboard_outbox
		order by id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.EventType, &row.UserID, &row.ContestID, &row.Year, &row.Unprocessed); err != nil {
			t.Fatal(err)
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []outboxRow{
		{EventType: "refresh_official_scores", UserID: testUserID, Unprocessed: true},
		{EventType: "refresh_contest_score", UserID: testOtherUserID, ContestID: &contestID, Year: &year, Unprocessed: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("outbox rows=%+v, want %+v", got, want)
	}
}

// Pins the legacy array-text tag decoding, including its known mishandling of
// escaped quotes and backslashes, until that response contract is reviewed.
func TestLogsRepositoryLegacyTagDecoding(t *testing.T) {
	t.Parallel()
	repository, _ := newTestLogsRepository(t)

	logID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1")
	createDurationLog(t, repository, logID, testUserID, time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))

	// Digit prefixes keep array_agg ordering independent of collation.
	for _, tag := range []string{`1 "quoted"`, `2,comma`, `3\back`, `4plain`, `5a",b`} {
		if err := repository.InsertTag(t.Context(), logID, testUserID, tag); err != nil {
			t.Fatal(err)
		}
	}

	found, err := repository.FindLog(t.Context(), logID, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{`1 \quoted\`, `2,comma`, `3\\back`, `4plain`, `5a\`, `b`}
	if !reflect.DeepEqual(found.Tags, want) {
		t.Errorf("tags=%q, want %q", found.Tags, want)
	}
}

package scoring

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestScoringRepositoryPlatformRuleSetLifecycle(t *testing.T) {
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

	repository := NewScoringRepository(db.Pool)
	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	seededActive, err := repository.FindActivePlatformRuleSet(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	seeded, err := repository.ListPlatformRuleSets(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(seeded) == 0 || seeded[0].ID != seededActive.ID {
		t.Fatalf("seeded platform rule sets=%+v, want newest first with active %s", seeded, seededActive.ID)
	}
	for i := 1; i < len(seeded); i++ {
		if seeded[i-1].Version <= seeded[i].Version {
			t.Errorf("platform rule sets not ordered by version desc: %d before %d", seeded[i-1].Version, seeded[i].Version)
		}
	}

	version, err := repository.NextPlatformDraftVersion(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if version != seeded[0].Version+1 {
		t.Errorf("next platform version=%d, want %d", version, seeded[0].Version+1)
	}

	draft, err := repository.CreateDraft(t.Context(), RuleSet{
		ID:        uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"),
		Scope:     "platform",
		Version:   version,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantDraft := &RuleSet{
		ID:      uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"),
		Scope:   "platform",
		Version: version,
		Status:  "draft",
		Rules:   []Rule{},
	}
	if !draft.CreatedAt.Equal(createdAt) {
		t.Errorf("draft created_at=%s, want %s", draft.CreatedAt, createdAt)
	}
	wantDraft.CreatedAt = draft.CreatedAt
	if !reflect.DeepEqual(draft, wantDraft) {
		t.Errorf("draft=%+v, want %+v", draft, wantDraft)
	}

	_, err = repository.CreateDraft(t.Context(), RuleSet{
		ID:        uuid.New(),
		Scope:     "platform",
		Version:   version,
		CreatedAt: createdAt,
	})
	assertConstraintViolation(t, err, pgerrcode.UniqueViolation, "scoring_rule_sets_platform_version")

	specific := Rule{
		ID:           uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1"),
		Priority:     2,
		Stackable:    true,
		ActivityID:   1,
		UnitKey:      "reading_page",
		LanguageCode: "jpn",
		Tag:          "ebook",
		Source:       SourceAmount,
		Rate:         1.5,
	}
	general := Rule{
		ID:         uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2"),
		Priority:   1,
		ActivityID: 2,
		Source:     SourceDurationMinutes,
		Rate:       0.25,
	}
	for _, rule := range []Rule{specific, general} {
		if err := repository.CreateRule(t.Context(), draft.ID, rule); err != nil {
			t.Fatal(err)
		}
	}

	duplicate := general
	duplicate.ID = uuid.New()
	err = repository.CreateRule(t.Context(), draft.ID, duplicate)
	assertConstraintViolation(t, err, pgerrcode.UniqueViolation, "scoring_rules_rule_set_priority")

	mismatchedUnit := general
	mismatchedUnit.ID = uuid.New()
	mismatchedUnit.Priority = 3
	mismatchedUnit.UnitKey = "reading_page"
	err = repository.CreateRule(t.Context(), draft.ID, mismatchedUnit)
	assertConstraintViolation(t, err, pgerrcode.CheckViolation, "scoring_rules_unit_key_valid")

	rules, err := repository.ListRules(t.Context(), draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Rule{general, specific}; !reflect.DeepEqual(rules, want) {
		t.Errorf("rules=%+v, want %+v", rules, want)
	}

	published, err := repository.PublishRuleSet(t.Context(), draft.ID, publishedAt)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != "published" || published.PublishedAt == nil || !published.PublishedAt.Equal(publishedAt) {
		t.Errorf("published=%+v, want published at %s", published, publishedAt)
	}

	if _, err := repository.PublishRuleSet(t.Context(), draft.ID, publishedAt); errx.KindOf(err) != errx.Conflict {
		t.Errorf("republish error=%v, want conflict", err)
	}
	if _, err := repository.PublishRuleSet(t.Context(), uuid.New(), publishedAt); errx.KindOf(err) != errx.Conflict {
		t.Errorf("publish missing error=%v, want conflict", err)
	}

	if err := repository.ActivatePlatformRuleSet(t.Context(), draft.ID); err != nil {
		t.Fatal(err)
	}
	active, err := repository.FindActivePlatformRuleSet(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != draft.ID {
		t.Errorf("active platform rule set=%s, want %s", active.ID, draft.ID)
	}

	found, err := repository.FindRuleSetByID(t.Context(), seededActive.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Version != seededActive.Version || found.Status != "published" {
		t.Errorf("previous platform rule set=%+v, want unchanged published version %d", found, seededActive.Version)
	}

	if _, err := repository.FindRuleSetByID(t.Context(), uuid.New()); !errors.Is(err, ErrRuleSetNotFound) {
		t.Errorf("missing rule set error=%v, want not found", err)
	}
}

func TestScoringRepositoryContestRuleSets(t *testing.T) {
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
		values
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111', 'Owner', false,
			 '2026-10-01', '2026-10-31', '2026-10-15', 'Live', '{1}', false,
			 '2026-09-01', '2026-09-01', null),
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111', 'Owner', false,
			 '2026-10-01', '2026-10-31', '2026-10-15', 'Deleted', '{1}', false,
			 '2026-09-01', '2026-09-01', '2026-09-02')`)
	if err != nil {
		t.Fatal(err)
	}

	repository := NewScoringRepository(db.Pool)
	contestID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc1")
	deletedContestID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-ccccccccccc2")
	createdAt := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)

	activeID, err := repository.FindContestActiveRuleSetID(t.Context(), contestID)
	if err != nil || activeID != nil {
		t.Errorf("unconfigured contest rule set=%v err=%v, want nil", activeID, err)
	}
	for _, id := range []uuid.UUID{deletedContestID, uuid.New()} {
		if _, err := repository.FindContestActiveRuleSetID(t.Context(), id); !errors.Is(err, ErrContestNotFound) {
			t.Errorf("contest %s error=%v, want not found", id, err)
		}
	}

	version, err := repository.NextContestDraftVersion(t.Context(), contestID)
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("first contest version=%d, want 1", version)
	}

	replace, err := repository.CreateDraft(t.Context(), RuleSet{
		ID:        uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd1"),
		Scope:     "contest",
		ContestID: &contestID,
		Version:   1,
		Mode:      string(ModeReplace),
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	override, err := repository.CreateDraft(t.Context(), RuleSet{
		ID:                uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd2"),
		Scope:             "contest",
		ContestID:         &contestID,
		Version:           2,
		Mode:              string(ModeOverride),
		FallbackRuleSetID: &replace.ID,
		CreatedAt:         createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.CreateDraft(t.Context(), RuleSet{
		ID:        uuid.New(),
		Scope:     "contest",
		ContestID: &contestID,
		Version:   3,
		Mode:      string(ModeOverride),
		CreatedAt: createdAt,
	})
	assertConstraintViolation(t, err, pgerrcode.CheckViolation, "scoring_rule_sets_ownership_valid")

	list, err := repository.ListContestRuleSets(t.Context(), contestID)
	if err != nil {
		t.Fatal(err)
	}
	if want := []RuleSet{*override, *replace}; !reflect.DeepEqual(list, want) {
		t.Errorf("contest rule sets=%+v, want %+v", list, want)
	}
	if override.FallbackRuleSetID == nil || *override.FallbackRuleSetID != replace.ID || override.Mode != string(ModeOverride) {
		t.Errorf("override=%+v, want fallback %s", override, replace.ID)
	}

	platform, err := repository.ListPlatformRuleSets(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleSet := range platform {
		if ruleSet.Scope != "platform" {
			t.Errorf("platform list contains %+v", ruleSet)
		}
	}

	version, err = repository.NextContestDraftVersion(t.Context(), contestID)
	if err != nil {
		t.Fatal(err)
	}
	if version != 3 {
		t.Errorf("next contest version=%d, want 3", version)
	}

	if _, err := repository.PublishRuleSet(t.Context(), replace.ID, updatedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.ActivateContestRuleSet(t.Context(), contestID, replace.ID, updatedAt); err != nil {
		t.Fatal(err)
	}
	activeID, err = repository.FindContestActiveRuleSetID(t.Context(), contestID)
	if err != nil {
		t.Fatal(err)
	}
	if activeID == nil || *activeID != replace.ID {
		t.Errorf("active contest rule set=%v, want %s", activeID, replace.ID)
	}

	if err := repository.ActivateContestRuleSet(t.Context(), deletedContestID, replace.ID, updatedAt); err != nil {
		t.Fatal(err)
	}

	var liveUpdatedAt time.Time
	var deletedUnconfigured bool
	if err := db.Pool.QueryRow(t.Context(), `
		select
			(select updated_at from contests where id = 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1'),
			(select scoring_rule_set_id is null from contests where id = 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2')`,
	).Scan(&liveUpdatedAt, &deletedUnconfigured); err != nil {
		t.Fatal(err)
	}
	if !liveUpdatedAt.Equal(updatedAt) {
		t.Errorf("live contest updated_at=%s, want %s", liveUpdatedAt, updatedAt)
	}
	if !deletedUnconfigured {
		t.Error("deleted contest scoring_rule_set_id changed, want unchanged null")
	}
}

func TestScoringRepositoryUnitLookups(t *testing.T) {
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

	seededUnit := func(key string, languageCode *string) logUnit {
		t.Helper()

		unit := logUnit{Key: key}
		err := db.Pool.QueryRow(t.Context(), `
			select id, modifier from log_units
			where unit_key = $1 and log_activity_id = 1 and language_code is not distinct from $2`,
			key, languageCode,
		).Scan(&unit.ID, &unit.Modifier)
		if err != nil {
			t.Fatal(err)
		}
		return unit
	}
	jpn := "jpn"
	fallback := seededUnit("reading_character", nil)
	japanese := seededUnit("reading_character", &jpn)
	jpnOnly := seededUnit("reading_two_column_page", &jpn)

	repository := NewScoringRepository(db.Pool)
	missingID := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")

	idTests := []struct {
		name         string
		id           uuid.UUID
		activityID   int32
		languageCode string
		want         *logUnit
	}{
		{name: "language specific", id: japanese.ID, activityID: 1, languageCode: "jpn", want: &japanese},
		{name: "fallback for any language", id: fallback.ID, activityID: 1, languageCode: "eng", want: &fallback},
		{name: "other language", id: japanese.ID, activityID: 1, languageCode: "eng"},
		{name: "other activity", id: fallback.ID, activityID: 2, languageCode: "jpn"},
		{name: "missing", id: missingID, activityID: 1, languageCode: "jpn"},
	}
	for _, tt := range idTests {
		t.Run("by id "+tt.name, func(t *testing.T) {
			key, err := repository.FindUnitKeyByID(t.Context(), tt.id, tt.activityID, tt.languageCode)
			unit, unitErr := repository.FindLogUnitByID(t.Context(), tt.id, tt.activityID, tt.languageCode)
			assertUnitLookup(t, key, err, unit, unitErr, tt.want)
		})
	}

	keyTests := []struct {
		name         string
		key          string
		activityID   int32
		languageCode string
		want         *logUnit
	}{
		{name: "prefers language specific", key: "reading_character", activityID: 1, languageCode: "jpn", want: &japanese},
		{name: "falls back without language match", key: "reading_character", activityID: 1, languageCode: "eng", want: &fallback},
		{name: "language specific only", key: "reading_two_column_page", activityID: 1, languageCode: "jpn", want: &jpnOnly},
		{name: "other language", key: "reading_two_column_page", activityID: 1, languageCode: "eng"},
		{name: "other activity", key: "reading_character", activityID: 2, languageCode: "jpn"},
		{name: "missing", key: "missing_unit", activityID: 1, languageCode: "jpn"},
	}
	for _, tt := range keyTests {
		t.Run("by key "+tt.name, func(t *testing.T) {
			key, err := repository.FindUnitKeyByKey(t.Context(), tt.key, tt.activityID, tt.languageCode)
			unit, unitErr := repository.FindLogUnitByKey(t.Context(), tt.key, tt.activityID, tt.languageCode)
			assertUnitLookup(t, key, err, unit, unitErr, tt.want)
		})
	}
}

func assertUnitLookup(t *testing.T, key string, keyErr error, unit *logUnit, unitErr error, want *logUnit) {
	t.Helper()

	if want == nil {
		if errx.KindOf(keyErr) != errx.InvalidInput || key != "" {
			t.Errorf("unit key=%q err=%v, want invalid input", key, keyErr)
		}
		if errx.KindOf(unitErr) != errx.InvalidInput || unit != nil {
			t.Errorf("log unit=%+v err=%v, want invalid input", unit, unitErr)
		}
		return
	}

	if keyErr != nil || key != want.Key {
		t.Errorf("unit key=%q err=%v, want %q", key, keyErr, want.Key)
	}
	if unitErr != nil || unit == nil || *unit != *want {
		t.Errorf("log unit=%+v err=%v, want %+v", unit, unitErr, *want)
	}
}

func assertConstraintViolation(t *testing.T, err error, code, constraint string) {
	t.Helper()

	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != code || pgError.ConstraintName != constraint {
		t.Errorf("error=%v, want %s violation of %s", err, code, constraint)
	}
}

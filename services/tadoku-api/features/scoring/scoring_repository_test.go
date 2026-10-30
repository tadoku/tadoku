package scoring

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestScoringRepositoryPlatformRuleSetLifecycle(t *testing.T) {
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

	repository := NewScoringRepository(db.Pool)
	createdAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	publishedAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)

	seededActive, err := repository.FindActivePlatformRuleSet(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}

	seeded, err := repository.ListPlatformRuleSets(tenantCtx)
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

	version, err := repository.NextPlatformDraftVersion(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	if version != seeded[0].Version+1 {
		t.Errorf("next platform version=%d, want %d", version, seeded[0].Version+1)
	}

	draft, err := repository.CreateDraft(tenantCtx, RuleSet{
		ID:        uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"),
		Scope:     "platform",
		Version:   version,
		Status:    "draft",
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

	_, err = repository.CreateDraft(tenantCtx, RuleSet{
		ID:        uuid.New(),
		Scope:     "platform",
		Version:   version,
		Status:    "draft",
		CreatedAt: createdAt,
	})
	assertConstraintViolation(
		t, err, pgerrcode.UniqueViolation,
		"scoring_rule_sets_tenant_platform_version",
	)

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
		if err := repository.CreateRule(tenantCtx, draft.ID, rule); err != nil {
			t.Fatal(err)
		}
	}

	duplicate := general
	duplicate.ID = uuid.New()
	err = repository.CreateRule(tenantCtx, draft.ID, duplicate)
	assertConstraintViolation(
		t, err, pgerrcode.UniqueViolation,
		"scoring_rules_tenant_rule_set_priority",
	)

	mismatchedUnit := general
	mismatchedUnit.ID = uuid.New()
	mismatchedUnit.Priority = 3
	mismatchedUnit.UnitKey = "reading_page"
	err = repository.CreateRule(tenantCtx, draft.ID, mismatchedUnit)
	assertConstraintViolation(t, err, pgerrcode.CheckViolation, "scoring_rules_unit_key_valid")

	rules, err := repository.ListRules(tenantCtx, draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := []Rule{general, specific}; !reflect.DeepEqual(rules, want) {
		t.Errorf("rules=%+v, want %+v", rules, want)
	}

	published, err := repository.PublishRuleSet(tenantCtx, draft.ID, publishedAt)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != "published" || published.PublishedAt == nil || !published.PublishedAt.Equal(publishedAt) {
		t.Errorf("published=%+v, want published at %s", published, publishedAt)
	}

	_, err = repository.PublishRuleSet(tenantCtx, draft.ID, publishedAt)
	if !errors.Is(err, ErrRuleSetNotDraft) {
		t.Errorf("republish error=%v, want %v", err, ErrRuleSetNotDraft)
	}

	_, err = repository.PublishRuleSet(tenantCtx, uuid.New(), publishedAt)
	if !errors.Is(err, ErrRuleSetNotDraft) {
		t.Errorf("publish missing error=%v, want %v", err, ErrRuleSetNotDraft)
	}

	if err := repository.ActivatePlatformRuleSet(tenantCtx, draft.ID); err != nil {
		t.Fatal(err)
	}
	active, err := repository.FindActivePlatformRuleSet(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != draft.ID {
		t.Errorf("active platform rule set=%s, want %s", active.ID, draft.ID)
	}

	found, err := repository.FindRuleSetByID(tenantCtx, seededActive.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.Version != seededActive.Version || found.Status != "published" {
		t.Errorf("previous platform rule set=%+v, want unchanged published version %d", found, seededActive.Version)
	}

	if _, err := repository.FindRuleSetByID(tenantCtx, uuid.New()); !errors.Is(err, ErrRuleSetNotFound) {
		t.Errorf("missing rule set error=%v, want not found", err)
	}
}

func TestScoringRepositoryContestRuleSets(t *testing.T) {
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
		insert into contests (
			tenant, id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
			registration_end, title, activity_type_id_allow_list, official, created_at, updated_at, deleted_at
		)
		values
			('tadoku/prod', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111', 'Owner', false,
			 '2026-10-01', '2026-10-31', '2026-10-15', 'Live', '{1}', false,
			 '2026-09-01', '2026-09-01', null),
			('tadoku/prod', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111', 'Owner', false,
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

	activeID, err := repository.FindContestActiveRuleSetID(tenantCtx, contestID)
	if err != nil || activeID != nil {
		t.Errorf("unconfigured contest rule set=%v err=%v, want nil", activeID, err)
	}
	for _, id := range []uuid.UUID{deletedContestID, uuid.New()} {
		if _, err := repository.FindContestActiveRuleSetID(tenantCtx, id); !errors.Is(err, ErrContestNotFound) {
			t.Errorf("contest %s error=%v, want not found", id, err)
		}
	}

	version, err := repository.NextContestDraftVersion(tenantCtx, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("first contest version=%d, want 1", version)
	}

	replace, err := repository.CreateDraft(tenantCtx, RuleSet{
		ID:        uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd1"),
		Scope:     "contest",
		ContestID: &contestID,
		Version:   1,
		Status:    "draft",
		Mode:      string(modeReplace),
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	override, err := repository.CreateDraft(tenantCtx, RuleSet{
		ID:                uuid.MustParse("dddddddd-dddd-4ddd-8ddd-ddddddddddd2"),
		Scope:             "contest",
		ContestID:         &contestID,
		Version:           2,
		Status:            "draft",
		Mode:              string(modeOverride),
		FallbackRuleSetID: &replace.ID,
		CreatedAt:         createdAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = repository.CreateDraft(tenantCtx, RuleSet{
		ID:        uuid.New(),
		Scope:     "contest",
		ContestID: &contestID,
		Version:   3,
		Status:    "draft",
		Mode:      string(modeOverride),
		CreatedAt: createdAt,
	})
	assertConstraintViolation(t, err, pgerrcode.CheckViolation, "scoring_rule_sets_ownership_valid")

	list, err := repository.ListContestRuleSets(tenantCtx, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if want := []RuleSet{*override, *replace}; !reflect.DeepEqual(list, want) {
		t.Errorf("contest rule sets=%+v, want %+v", list, want)
	}
	if override.FallbackRuleSetID == nil || *override.FallbackRuleSetID != replace.ID || override.Mode != string(modeOverride) {
		t.Errorf("override=%+v, want fallback %s", override, replace.ID)
	}

	platform, err := repository.ListPlatformRuleSets(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleSet := range platform {
		if ruleSet.Scope != "platform" {
			t.Errorf("platform list contains %+v", ruleSet)
		}
	}

	version, err = repository.NextContestDraftVersion(tenantCtx, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if version != 3 {
		t.Errorf("next contest version=%d, want 3", version)
	}

	if _, err := repository.PublishRuleSet(tenantCtx, replace.ID, updatedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.ActivateContestRuleSet(tenantCtx, contestID, replace.ID, updatedAt); err != nil {
		t.Fatal(err)
	}
	activeID, err = repository.FindContestActiveRuleSetID(tenantCtx, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if activeID == nil || *activeID != replace.ID {
		t.Errorf("active contest rule set=%v, want %s", activeID, replace.ID)
	}

	if err := repository.ActivateContestRuleSet(tenantCtx, deletedContestID, replace.ID, updatedAt); !errors.Is(err, ErrContestNotFound) {
		t.Errorf("deleted contest activate error=%v, want contest not found", err)
	}

	var liveUpdatedAt time.Time
	var deletedUnconfigured bool
	if err := db.Pool.QueryRow(tenantCtx, `
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

	seededUnit := func(key string, languageCode *string) logUnit {
		t.Helper()

		unit := logUnit{Key: key}
		err := db.Pool.QueryRow(tenantCtx, `
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
	fallback := unitCandidate{logUnit: seededUnit("reading_character", nil), activityID: 1}
	characters := []unitCandidate{fallback}
	for _, code := range []string{"jpn", "kor", "zho", "cmn", "yue", "wuu"} {
		unit := seededUnit("reading_character", &code)
		characters = append(characters, unitCandidate{logUnit: unit, activityID: 1, languageCode: &code})
	}
	japanese := characters[1]
	jpnOnly := unitCandidate{logUnit: seededUnit("reading_two_column_page", &jpn), activityID: 1, languageCode: &jpn}

	repository := NewScoringRepository(db.Pool)
	missingID := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")

	idTests := []struct {
		name string
		id   uuid.UUID
		want []unitCandidate
	}{
		{name: "language specific", id: japanese.ID, want: []unitCandidate{japanese}},
		{name: "without language", id: fallback.ID, want: []unitCandidate{fallback}},
		{name: "missing", id: missingID},
	}
	for _, tt := range idTests {
		t.Run("by id "+tt.name, func(t *testing.T) {
			candidates, err := repository.ListUnitsByID(tenantCtx, tt.id)

			assertUnitCandidates(t, candidates, err, tt.want)
		})
	}

	keyTests := []struct {
		name string
		key  string
		want []unitCandidate
	}{
		{name: "with and without language", key: "reading_character", want: characters},
		{name: "language specific only", key: "reading_two_column_page", want: []unitCandidate{jpnOnly}},
		{name: "missing", key: "missing_unit"},
	}
	for _, tt := range keyTests {
		t.Run("by key "+tt.name, func(t *testing.T) {
			candidates, err := repository.ListUnitsByKey(tenantCtx, tt.key)

			assertUnitCandidates(t, candidates, err, tt.want)
		})
	}
}

func assertUnitCandidates(t *testing.T, candidates []unitCandidate, err error, want []unitCandidate) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != len(want) {
		t.Fatalf("unit candidates=%+v, want %d candidates", candidates, len(want))
	}
	for _, wanted := range want {
		matches := func(candidate unitCandidate) bool { return reflect.DeepEqual(candidate, wanted) }
		if !slices.ContainsFunc(candidates, matches) {
			t.Errorf("unit candidates=%+v, want %+v with language %v", candidates, wanted.logUnit, wanted.languageCode)
		}
	}
}

func assertConstraintViolation(t *testing.T, err error, code string, constraints ...string) {
	t.Helper()

	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != code || !slices.Contains(constraints, pgError.ConstraintName) {
		t.Errorf("error=%v, want %s violation of %v", err, code, constraints)
	}
}

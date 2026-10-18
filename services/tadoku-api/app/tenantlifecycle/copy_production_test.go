package tenantlifecycle_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func seedProductionPages(t *testing.T, db *testpostgres.Database) {
	t.Helper()

	_, err := db.Pool.Exec(t.Context(), `with seeded_pages as (
			insert into pages (tenant, id, namespace, slug, current_content_id, published_at, deleted_at)
			values
				('tadoku/prod', uuid_generate_v5(uuid_nil(), 'about'), 'tadoku', 'about',
					uuid_generate_v5(uuid_nil(), 'about/2'), now() at time zone 'utc' - interval '1 day', null),
				('tadoku/prod', uuid_generate_v5(uuid_nil(), 'draft'), 'tadoku', 'draft',
					uuid_generate_v5(uuid_nil(), 'draft/1'), null, null),
				('tadoku/prod', uuid_generate_v5(uuid_nil(), 'scheduled'), 'tadoku', 'scheduled',
					uuid_generate_v5(uuid_nil(), 'scheduled/1'), now() at time zone 'utc' + interval '1 day', null),
				('tadoku/prod', uuid_generate_v5(uuid_nil(), 'deleted'), 'tadoku', 'deleted',
					uuid_generate_v5(uuid_nil(), 'deleted/1'), now() at time zone 'utc' - interval '1 day', now())
		)
		insert into pages_content (tenant, id, page_id, title, html, created_at)
		values
			('tadoku/prod', uuid_generate_v5(uuid_nil(), 'about/1'), uuid_generate_v5(uuid_nil(), 'about'),
				'Old about', '<p>Old</p>', now() - interval '2 days'),
			('tadoku/prod', uuid_generate_v5(uuid_nil(), 'about/2'), uuid_generate_v5(uuid_nil(), 'about'),
				'About', '<p>Current</p>', now() - interval '1 day'),
			('tadoku/prod', uuid_generate_v5(uuid_nil(), 'draft/1'), uuid_generate_v5(uuid_nil(), 'draft'),
				'Draft', '<p>Draft</p>', now()),
			('tadoku/prod', uuid_generate_v5(uuid_nil(), 'scheduled/1'), uuid_generate_v5(uuid_nil(), 'scheduled'),
				'Scheduled', '<p>Scheduled</p>', now()),
			('tadoku/prod', uuid_generate_v5(uuid_nil(), 'deleted/1'), uuid_generate_v5(uuid_nil(), 'deleted'),
				'Deleted', '<p>Deleted</p>', now())`)
	if err != nil {
		t.Fatal(err)
	}
}

func assertCopiedPublishedPages(t *testing.T, db *testpostgres.Database, key tenant.TestKey) {
	t.Helper()
	ctx := tenant.WithKey(t.Context(), key.Key())

	service := pages.NewService(pages.NewPagesRepository(db.AppPool))
	page, err := service.FindPageBySlug(ctx, "tadoku", "about")
	if err != nil || page.Title != "About" || page.HTML != "<p>Current</p>" {
		t.Fatalf("about page in provisioned tenant: page=%+v error=%v", page, err)
	}

	var production uuid.UUID
	var copiedPages, copiedContents int
	err = db.Pool.QueryRow(
		t.Context(),
		`select
			uuid_generate_v5(uuid_nil(), 'about'),
			(select count(*) from pages where tenant = $1),
			(select count(*) from pages_content where tenant = $1)`,
		key.String(),
	).Scan(&production, &copiedPages, &copiedContents)
	if err != nil {
		t.Fatal(err)
	}
	if page.ID == production || copiedPages != 1 || copiedContents != 1 {
		t.Errorf(
			"copied pages: about=%s production=%s pages=%d contents=%d, want one page and its current revision",
			page.ID,
			production,
			copiedPages,
			copiedContents,
		)
	}
}

func assertCopiedPlatformScoring(t *testing.T, db *testpostgres.Database, key tenant.TestKey) {
	t.Helper()
	ctx := tenant.WithKey(t.Context(), key.Key())

	service := scoring.NewService(scoring.NewScoringRepository(db.AppPool), true, nil)
	unit := "reading_page"
	amount := float32(10)
	estimate, eligible, err := service.ScorePlatform(ctx, scoring.PreviewParameters{
		UnitKey:      &unit,
		ActivityID:   1,
		LanguageCode: "jpn",
		Amount:       &amount,
	})
	if err != nil || !eligible || estimate.Score != 10 || estimate.RuleSetID == nil {
		t.Fatalf("score in provisioned tenant: estimate=%+v eligible=%t error=%v", estimate, eligible, err)
	}

	var production uuid.UUID
	var sets, platformSets, differing int
	err = db.Pool.QueryRow(t.Context(), `with copied as (
			select active_rule_set_id as id from platform_scoring_config where tenant = $1
		),
		canonical as (
			select active_rule_set_id as id from platform_scoring_config where tenant = 'tadoku/prod'
		),
		copied_rules as (
			select priority, stackable, activity_id, unit_key, language_code, tag, score_source, rate
			from scoring_rules
			where rule_set_id = (select id from copied)
				and tenant = $1
		),
		canonical_rules as (
			select priority, stackable, activity_id, unit_key, language_code, tag, score_source, rate
			from scoring_rules
			where rule_set_id = (select id from canonical)
		)
		select
			(select id from canonical),
			(select count(*) from scoring_rule_sets where tenant = $1),
			(select count(*) from scoring_rule_sets where tenant = $1 and scope = 'platform'),
			(select count(*) from (
				(table copied_rules except all table canonical_rules)
				union all (table canonical_rules except all table copied_rules)
			) difference)`,
		key.String(),
	).Scan(&production, &sets, &platformSets, &differing)
	if err != nil {
		t.Fatal(err)
	}
	if *estimate.RuleSetID == production || sets != 1 || platformSets != 1 || differing != 0 {
		t.Errorf(
			"copied scoring: rule set=%s production=%s sets=%d platform=%d differing rules=%d",
			*estimate.RuleSetID,
			production,
			sets,
			platformSets,
			differing,
		)
	}
}

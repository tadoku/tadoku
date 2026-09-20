package languages_test

import (
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestLanguagesRepositoryListsByName(t *testing.T) {
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
	if err := db.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Pool.Exec(t.Context(), `
		insert into languages (code, name)
		values ('test-z', 'Zulu test'), ('test-a', 'Ainu test')`); err != nil {
		t.Fatal(err)
	}

	items, err := languages.NewLanguagesRepository(db.Pool).ListLanguages(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 {
		t.Fatalf("languages=%d, want at least 2", len(items))
	}
	if got := items[0]; got != (languages.Language{Code: "afr", Name: "Afrikaans"}) {
		t.Errorf("first language=%+v, want Afrikaans", got)
	}
	if got := items[len(items)-1]; got != (languages.Language{Code: "test-z", Name: "Zulu test"}) {
		t.Errorf("last language=%+v, want Zulu test", got)
	}

	if _, err := db.Pool.Exec(t.Context(), "delete from scoring_rules; delete from languages"); err != nil {
		t.Fatal(err)
	}
	items, err = languages.NewLanguagesRepository(db.Pool).ListLanguages(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 {
		t.Errorf("empty languages=%v, want non-nil empty slice", items)
	}
}

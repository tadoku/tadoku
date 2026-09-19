package languages_test

import (
	"errors"
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

func TestLanguagesRepositoryCreatesLanguageAndRejectsDuplicateCode(t *testing.T) {
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

	repository := languages.NewLanguagesRepository(db.Pool)
	parameters := languages.CreateLanguageParameters{Code: "test-new", Name: "  Test language  "}
	if err := repository.CreateLanguage(t.Context(), parameters); err != nil {
		t.Fatal(err)
	}

	var name string
	if err := db.Pool.QueryRow(t.Context(), "select name from languages where code = $1", parameters.Code).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != parameters.Name {
		t.Errorf("name=%q, want exact %q", name, parameters.Name)
	}
	duplicate := parameters
	duplicate.Name = "Replacement"
	if err := repository.CreateLanguage(t.Context(), duplicate); !errors.Is(err, languages.ErrLanguageAlreadyExists) {
		t.Errorf("duplicate error=%v, want language already exists", err)
	}
	if err := db.Pool.QueryRow(t.Context(), "select name from languages where code = $1", parameters.Code).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != parameters.Name {
		t.Errorf("name after conflict=%q, want original %q", name, parameters.Name)
	}
}

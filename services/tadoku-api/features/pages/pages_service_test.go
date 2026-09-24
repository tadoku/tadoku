package pages_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreatePageParametersValidation(t *testing.T) {
	t.Parallel()
	valid := pages.CreatePageParameters{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Slug:      "new-page",
		Title:     "Title",
		HTML:      "<p>Page</p>",
	}
	for _, test := range []struct {
		name    string
		change  func(*pages.CreatePageParameters)
		message string
	}{
		{name: "ID", change: func(p *pages.CreatePageParameters) { p.ID = uuid.Nil }, message: "id is required"},
		{name: "namespace", change: func(p *pages.CreatePageParameters) { p.Namespace = "" }, message: "namespace is required"},
		{name: "short slug", change: func(p *pages.CreatePageParameters) { p.Slug = "x" }, message: "slug must be at least 2 characters"},
		{name: "uppercase slug", change: func(p *pages.CreatePageParameters) { p.Slug = "New-page" }, message: "slug must be lowercase"},
		{name: "title", change: func(p *pages.CreatePageParameters) { p.Title = "" }, message: "title is required"},
		{name: "HTML", change: func(p *pages.CreatePageParameters) { p.HTML = "" }, message: "html is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			err := parameters.Validate()
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != test.message {
				t.Errorf("error=%v, want invalid input %q", err, test.message)
			}
		})
	}

	valid.Slug = "日本"
	valid.Title = " "
	valid.HTML = "\t"
	if err := valid.Validate(); err != nil {
		t.Errorf("valid whitespace and Unicode fields rejected: %v", err)
	}
}

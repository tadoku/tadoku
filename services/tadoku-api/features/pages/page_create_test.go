package pages_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
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
		name   string
		change func(*pages.CreatePageParameters)
	}{
		{name: "ID", change: func(p *pages.CreatePageParameters) { p.ID = uuid.Nil }},
		{name: "namespace", change: func(p *pages.CreatePageParameters) { p.Namespace = "" }},
		{name: "short slug", change: func(p *pages.CreatePageParameters) { p.Slug = "x" }},
		{name: "uppercase slug", change: func(p *pages.CreatePageParameters) { p.Slug = "New-page" }},
		{name: "title", change: func(p *pages.CreatePageParameters) { p.Title = "" }},
		{name: "HTML", change: func(p *pages.CreatePageParameters) { p.HTML = "" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			if err := parameters.Validate(); !errors.Is(err, pages.ErrInvalidPage) {
				t.Errorf("error=%v, want invalid page", err)
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

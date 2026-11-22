package pages_test

import (
	"strings"
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
		{name: "long namespace", change: func(p *pages.CreatePageParameters) { p.Namespace = strings.Repeat("a", 51) }, message: "namespace must be at most 50 characters"},
		{name: "short slug", change: func(p *pages.CreatePageParameters) { p.Slug = "x" }, message: "slug must be at least 2 characters"},
		{name: "long slug", change: func(p *pages.CreatePageParameters) { p.Slug = strings.Repeat("a", 201) }, message: "slug must be at most 200 characters"},
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

	valid.Slug = strings.Repeat("a", 200)
	if err := valid.Validate(); err != nil {
		t.Errorf("200-rune slug rejected: %v", err)
	}

	valid.Namespace = strings.Repeat("a", 50)
	if err := valid.Validate(); err != nil {
		t.Errorf("50-rune namespace rejected: %v", err)
	}
}

func TestUpdatePageParametersValidation(t *testing.T) {
	t.Parallel()
	valid := pages.UpdatePageParameters{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Slug:      "new-page",
		Title:     "Title",
		HTML:      "<p>Page</p>",
	}
	for _, test := range []struct {
		name    string
		change  func(*pages.UpdatePageParameters)
		message string
	}{
		{name: "ID", change: func(p *pages.UpdatePageParameters) { p.ID = uuid.Nil }, message: "id is required"},
		{name: "namespace", change: func(p *pages.UpdatePageParameters) { p.Namespace = "" }, message: "namespace is required"},
		{name: "long namespace", change: func(p *pages.UpdatePageParameters) { p.Namespace = strings.Repeat("a", 51) }, message: "namespace must be at most 50 characters"},
		{name: "short slug", change: func(p *pages.UpdatePageParameters) { p.Slug = "x" }, message: "slug must be at least 2 characters"},
		{name: "long slug", change: func(p *pages.UpdatePageParameters) { p.Slug = strings.Repeat("a", 201) }, message: "slug must be at most 200 characters"},
		{name: "uppercase slug", change: func(p *pages.UpdatePageParameters) { p.Slug = "New-page" }, message: "slug must be lowercase"},
		{name: "title", change: func(p *pages.UpdatePageParameters) { p.Title = "" }, message: "title is required"},
		{name: "HTML", change: func(p *pages.UpdatePageParameters) { p.HTML = "" }, message: "html is required"},
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

	if err := valid.Validate(); err != nil {
		t.Errorf("valid parameters rejected: %v", err)
	}

	valid.Slug = strings.Repeat("a", 200)
	if err := valid.Validate(); err != nil {
		t.Errorf("200-rune slug rejected: %v", err)
	}

	valid.Namespace = strings.Repeat("a", 50)
	if err := valid.Validate(); err != nil {
		t.Errorf("50-rune namespace rejected: %v", err)
	}
}

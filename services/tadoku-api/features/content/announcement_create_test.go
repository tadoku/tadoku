package content_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreateAnnouncementParametersValidation(t *testing.T) {
	t.Parallel()
	valid := content.CreateAnnouncementParameters{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Title:     "Title",
		Content:   "Content",
		Style:     "info",
		StartsAt:  time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	}
	for _, style := range []string{"info", "success", "warning", "error"} {
		t.Run(style, func(t *testing.T) {
			parameters := valid
			parameters.Style = style
			if err := parameters.Validate(); err != nil {
				t.Errorf("valid parameters rejected: %v", err)
			}
		})
	}
	for _, test := range []struct {
		name   string
		change func(*content.CreateAnnouncementParameters)
	}{
		{name: "ID", change: func(p *content.CreateAnnouncementParameters) { p.ID = uuid.Nil }},
		{name: "namespace", change: func(p *content.CreateAnnouncementParameters) { p.Namespace = "" }},
		{name: "title", change: func(p *content.CreateAnnouncementParameters) { p.Title = "" }},
		{name: "content", change: func(p *content.CreateAnnouncementParameters) { p.Content = "" }},
		{name: "style", change: func(p *content.CreateAnnouncementParameters) { p.Style = "other" }},
		{name: "empty style", change: func(p *content.CreateAnnouncementParameters) { p.Style = "" }},
		{name: "style case", change: func(p *content.CreateAnnouncementParameters) { p.Style = "INFO" }},
		{name: "starts at", change: func(p *content.CreateAnnouncementParameters) { p.StartsAt = time.Time{} }},
		{name: "ends at", change: func(p *content.CreateAnnouncementParameters) { p.EndsAt = time.Time{} }},
		{name: "equal dates", change: func(p *content.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt }},
		{name: "reversed dates", change: func(p *content.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt.Add(-time.Second) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			err := parameters.Validate()
			if !errors.Is(err, content.ErrInvalidAnnouncement) {
				t.Errorf("error=%v, want invalid announcement", err)
			}
			if errx.KindOf(err) != errx.InvalidInput {
				t.Errorf("error=%v, want invalid-input category", err)
			}
		})
	}
}

func TestCreateAnnouncementParametersHrefValidation(t *testing.T) {
	t.Parallel()
	parameters := content.CreateAnnouncementParameters{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Title:     "Title",
		Content:   "Content",
		Style:     "info",
		StartsAt:  time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	}

	for _, href := range []string{"", "/news", "https://example.test/news", "http://example.test", "/" + strings.Repeat("a", 2047)} {
		t.Run("valid "+href, func(t *testing.T) {
			parameters := parameters
			parameters.Href = &href
			if err := parameters.Validate(); err != nil {
				t.Errorf("valid href %q rejected: %v", href, err)
			}
		})
	}

	for _, href := range []string{"news", "//example.test/news", `/\example.test/news`, "javascript:alert(1)", "data:text/html,<script>alert(1)</script>", "vbscript:msgbox(1)", "https://example.test/%gh", "/" + strings.Repeat("a", 2048)} {
		t.Run("invalid "+href, func(t *testing.T) {
			parameters := parameters
			parameters.Href = &href
			if err := parameters.Validate(); !errors.Is(err, content.ErrInvalidAnnouncement) {
				t.Errorf("error=%v, want invalid announcement for href %q", err, href)
			}
		})
	}
}

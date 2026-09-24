package announcements_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestEmptyNamespaceIsRejectedBeforeStorage(t *testing.T) {
	service := announcements.NewService(nil)
	_, err := service.ListActiveAnnouncements(context.Background(), "")
	if errx.KindOf(err) != errx.InvalidInput || err.Error() != "namespace is required" {
		t.Errorf("error=%v, want invalid input %q", err, "namespace is required")
	}
}

func TestCreateAnnouncementParametersValidation(t *testing.T) {
	t.Parallel()
	valid := announcements.CreateAnnouncementParameters{
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
		name    string
		change  func(*announcements.CreateAnnouncementParameters)
		message string
	}{
		{name: "ID", change: func(p *announcements.CreateAnnouncementParameters) { p.ID = uuid.Nil }, message: "id is required"},
		{name: "namespace", change: func(p *announcements.CreateAnnouncementParameters) { p.Namespace = "" }, message: "namespace is required"},
		{name: "title", change: func(p *announcements.CreateAnnouncementParameters) { p.Title = "" }, message: "title is required"},
		{name: "content", change: func(p *announcements.CreateAnnouncementParameters) { p.Content = "" }, message: "content is required"},
		{name: "style", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "other" }, message: "style must be one of success, warning, error or info"},
		{name: "empty style", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "" }, message: "style must be one of success, warning, error or info"},
		{name: "style case", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "INFO" }, message: "style must be one of success, warning, error or info"},
		{name: "starts at", change: func(p *announcements.CreateAnnouncementParameters) { p.StartsAt = time.Time{} }, message: "starts_at is required"},
		{name: "ends at", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = time.Time{} }, message: "ends_at is required"},
		{name: "equal dates", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt }, message: "ends_at must be after starts_at"},
		{name: "reversed dates", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt.Add(-time.Second) }, message: "ends_at must be after starts_at"},
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
}

func TestCreateAnnouncementParametersHrefValidation(t *testing.T) {
	t.Parallel()
	parameters := announcements.CreateAnnouncementParameters{
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

	for _, test := range []struct {
		href    string
		message string
	}{
		{href: "news", message: "href must be an http(s) URL or a root-relative path"},
		{href: "//example.test/news", message: "href must be an http(s) URL or a root-relative path"},
		{href: `/\example.test/news`, message: "href must be an http(s) URL or a root-relative path"},
		{href: "javascript:alert(1)", message: "href must be an http(s) URL or a root-relative path"},
		{href: "data:text/html,<script>alert(1)</script>", message: "href must be an http(s) URL or a root-relative path"},
		{href: "vbscript:msgbox(1)", message: "href must be an http(s) URL or a root-relative path"},
		{href: "https://example.test/%gh", message: "href must be an http(s) URL or a root-relative path"},
		{href: "/" + strings.Repeat("a", 2048), message: "href must be at most 2048 characters"},
	} {
		t.Run("invalid "+test.href, func(t *testing.T) {
			parameters := parameters
			parameters.Href = &test.href
			err := parameters.Validate()
			if errx.KindOf(err) != errx.InvalidInput || err.Error() != test.message {
				t.Errorf("error=%v, want invalid input %q for href %q", err, test.message, test.href)
			}
		})
	}
}

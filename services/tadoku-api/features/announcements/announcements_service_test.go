package announcements_test

import (
	"context"
	"errors"
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
	if !errors.Is(err, announcements.ErrInvalidNamespace) {
		t.Errorf("error=%v want invalid namespace", err)
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
		name   string
		change func(*announcements.CreateAnnouncementParameters)
		want   string
	}{
		{name: "ID", change: func(p *announcements.CreateAnnouncementParameters) { p.ID = uuid.Nil }, want: "invalid announcement: id is nil"},
		{name: "namespace", change: func(p *announcements.CreateAnnouncementParameters) { p.Namespace = "" }, want: "invalid announcement: namespace is required"},
		{name: "title", change: func(p *announcements.CreateAnnouncementParameters) { p.Title = "" }, want: "invalid announcement: title is required"},
		{name: "content", change: func(p *announcements.CreateAnnouncementParameters) { p.Content = "" }, want: "invalid announcement: content is required"},
		{name: "style", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "other" }, want: "invalid announcement: style is invalid"},
		{name: "empty style", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "" }, want: "invalid announcement: style is invalid"},
		{name: "style case", change: func(p *announcements.CreateAnnouncementParameters) { p.Style = "INFO" }, want: "invalid announcement: style is invalid"},
		{name: "starts at", change: func(p *announcements.CreateAnnouncementParameters) { p.StartsAt = time.Time{} }, want: "invalid announcement: date range is invalid"},
		{name: "ends at", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = time.Time{} }, want: "invalid announcement: date range is invalid"},
		{name: "equal dates", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt }, want: "invalid announcement: date range is invalid"},
		{name: "reversed dates", change: func(p *announcements.CreateAnnouncementParameters) { p.EndsAt = p.StartsAt.Add(-time.Second) }, want: "invalid announcement: date range is invalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			err := parameters.Validate()
			if !errors.Is(err, announcements.ErrInvalidAnnouncement) {
				t.Errorf("error=%v, want invalid announcement", err)
			}
			if err == nil || err.Error() != test.want {
				t.Errorf("error=%v, want %q", err, test.want)
			}
			if errx.KindOf(err) != errx.InvalidInput {
				t.Errorf("error=%v, want invalid-input category", err)
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

	for _, href := range []string{"news", "//example.test/news", `/\example.test/news`, "javascript:alert(1)", "data:text/html,<script>alert(1)</script>", "vbscript:msgbox(1)", "https://example.test/%gh", "/" + strings.Repeat("a", 2048)} {
		t.Run("invalid "+href, func(t *testing.T) {
			parameters := parameters
			parameters.Href = &href
			err := parameters.Validate()
			if !errors.Is(err, announcements.ErrInvalidAnnouncement) {
				t.Errorf("error=%v, want invalid announcement for href %q", err, href)
			}
			if err == nil || err.Error() != "invalid announcement: href is invalid" {
				t.Errorf("error=%v, want href is invalid for href %q", err, href)
			}
		})
	}
}

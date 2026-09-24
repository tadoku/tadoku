package posts_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreatePostParametersValidation(t *testing.T) {
	t.Parallel()
	valid := posts.CreatePostParameters{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Slug:      "first-post",
		Title:     "Title",
		Content:   "Content",
	}
	for _, test := range []struct {
		name    string
		change  func(*posts.CreatePostParameters)
		message string
	}{
		{name: "ID", change: func(p *posts.CreatePostParameters) { p.ID = uuid.Nil }, message: "id is required"},
		{name: "namespace", change: func(p *posts.CreatePostParameters) { p.Namespace = "" }, message: "namespace is required"},
		{name: "slug", change: func(p *posts.CreatePostParameters) { p.Slug = "" }, message: "slug must be at least 2 characters"},
		{name: "one character slug", change: func(p *posts.CreatePostParameters) { p.Slug = "a" }, message: "slug must be at least 2 characters"},
		{name: "one Unicode character slug", change: func(p *posts.CreatePostParameters) { p.Slug = "日" }, message: "slug must be at least 2 characters"},
		{name: "uppercase slug", change: func(p *posts.CreatePostParameters) { p.Slug = "First-post" }, message: "slug must be lowercase"},
		{name: "uppercase Unicode slug", change: func(p *posts.CreatePostParameters) { p.Slug = "Éé" }, message: "slug must be lowercase"},
		{name: "title", change: func(p *posts.CreatePostParameters) { p.Title = "" }, message: "title is required"},
		{name: "content", change: func(p *posts.CreatePostParameters) { p.Content = "" }, message: "content is required"},
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

	for _, slug := range []string{"ab", "first-post", "日本", "éé", "12", "  "} {
		t.Run("valid slug "+slug, func(t *testing.T) {
			parameters := valid
			parameters.Slug = slug
			parameters.Title = " "
			parameters.Content = "\t"
			if err := parameters.Validate(); err != nil {
				t.Errorf("valid parameters rejected: %v", err)
			}
		})
	}
	for _, publishedAt := range []*time.Time{nil, new(time.Time)} {
		parameters := valid
		parameters.PublishedAt = publishedAt
		if err := parameters.Validate(); err != nil {
			t.Errorf("publication date %v rejected: %v", publishedAt, err)
		}
	}
}

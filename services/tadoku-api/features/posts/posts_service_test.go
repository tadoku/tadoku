package posts_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
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
		name   string
		change func(*posts.CreatePostParameters)
	}{
		{name: "ID", change: func(p *posts.CreatePostParameters) { p.ID = uuid.Nil }},
		{name: "namespace", change: func(p *posts.CreatePostParameters) { p.Namespace = "" }},
		{name: "slug", change: func(p *posts.CreatePostParameters) { p.Slug = "" }},
		{name: "one character slug", change: func(p *posts.CreatePostParameters) { p.Slug = "a" }},
		{name: "one Unicode character slug", change: func(p *posts.CreatePostParameters) { p.Slug = "日" }},
		{name: "uppercase slug", change: func(p *posts.CreatePostParameters) { p.Slug = "First-post" }},
		{name: "uppercase Unicode slug", change: func(p *posts.CreatePostParameters) { p.Slug = "Éé" }},
		{name: "title", change: func(p *posts.CreatePostParameters) { p.Title = "" }},
		{name: "content", change: func(p *posts.CreatePostParameters) { p.Content = "" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := valid
			test.change(&parameters)
			if err := parameters.Validate(); !errors.Is(err, posts.ErrInvalidPost) {
				t.Errorf("error=%v, want invalid post", err)
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

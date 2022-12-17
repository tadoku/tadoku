package pagefind

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrPageNotFound = errors.New("page not found")

type PageFindResponse struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Html        string
	PublishedAt *time.Time
}

type PageRepository interface {
	FindBySlug(context.Context, string) (*PageFindResponse, error)
}

type Service interface {
	FindBySlug(context.Context, string) (*PageFindResponse, error)
}

type service struct {
	pr PageRepository
}

func NewService(pr PageRepository) Service {
	return &service{
		pr: pr,
	}
}

func (s *service) FindBySlug(ctx context.Context, slug string) (*PageFindResponse, error) {
	page, err := s.pr.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// TODO: Extract out time.Now() into a clock provider so it can be mocked
	if page.PublishedAt == nil || page.PublishedAt.After(time.Now()) {
		return nil, fmt.Errorf("page is not published yet: %w", ErrPageNotFound)
	}

	return page, nil
}
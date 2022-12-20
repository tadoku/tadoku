package pagequery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrPageNotFound = errors.New("page not found")

type PageRepository interface {
	FindBySlug(context.Context, *PageFindRequest) (*PageFindResponse, error)
	ListPages(context.Context, *PageListRequest) (*PageListResponse, error)
}

type Service interface {
	FindBySlug(context.Context, *PageFindRequest) (*PageFindResponse, error)
	ListPages(context.Context, *PageListRequest) (*PageListResponse, error)
}

type service struct {
	pr PageRepository
}

func NewService(pr PageRepository) Service {
	return &service{
		pr: pr,
	}
}

type PageFindRequest struct {
	Slug      string `validate:"required"`
	Namespace string `validate:"required"`
}

type PageFindResponse struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	Html        string
	PublishedAt *time.Time
}

func (s *service) FindBySlug(ctx context.Context, req *PageFindRequest) (*PageFindResponse, error) {
	page, err := s.pr.FindBySlug(ctx, req)
	if err != nil {
		return nil, err
	}

	// TODO: Extract out time.Now() into a clock provider so it can be mocked
	if page.PublishedAt == nil || page.PublishedAt.After(time.Now()) {
		return nil, fmt.Errorf("page is not published yet: %w", ErrPageNotFound)
	}

	return page, nil
}

type PageListRequest struct {
	Namespace     string `validate:"required"`
	IncludeDrafts bool
	PageSize      int
	Page          int
}

type PageListResponse struct {
	Pages         []PageListEntry
	TotalSize     int
	NextPageToken string
}

type PageListEntry struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *service) ListPages(ctx context.Context, req *PageListRequest) (*PageListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.pr.ListPages(ctx, req)
}

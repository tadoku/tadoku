package pages

import (
	"context"
	"math"
	"strconv"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	pages *PagesRepository
}

func NewService(pages *PagesRepository) *Service {
	return &Service{pages: pages}
}

func (s *Service) CreatePage(ctx context.Context, parameters CreatePageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	now := timex.Now()
	item := &Page{
		ID:          parameters.ID,
		Namespace:   parameters.Namespace,
		Slug:        parameters.Slug,
		Title:       parameters.Title,
		HTML:        parameters.HTML,
		PublishedAt: parameters.PublishedAt,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	contentID := uuid.New()
	if err := s.pages.CreatePage(ctx, item, contentID); err != nil {
		return err
	}

	return s.pages.CreatePageContent(ctx, item.ID, contentID, item.Title, item.HTML, *item.CreatedAt)
}

func (s *Service) DeletePage(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.pages.DeletePage(ctx, namespace, id, timex.Now())
}

func (s *Service) FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error) {
	if namespace == "" {
		return nil, errx.NewInvalidInputError("namespace is required")
	}
	if slug == "" {
		return nil, errx.NewInvalidInputError("slug is required")
	}

	page, err := s.pages.FindPageBySlug(ctx, namespace, slug)
	if err != nil {
		return nil, err
	}
	if page.PublishedAt == nil || page.PublishedAt.After(timex.Now()) {
		return nil, ErrPageNotFound
	}

	return page, nil
}

func (s *Service) FindPageByID(ctx context.Context, namespace string, id uuid.UUID) (*Page, error) {
	return s.pages.FindPageByID(ctx, namespace, id)
}

func (s *Service) ListPages(ctx context.Context, namespace string, includeDrafts *bool, pageSize, page int) (*PageList, error) {
	if namespace == "" {
		return nil, errx.NewInvalidInputError("namespace is required")
	}
	if pageSize < 0 {
		return nil, errx.NewInvalidInputError("page_size must not be negative")
	}
	if page < 0 {
		return nil, errx.NewInvalidInputError("page must not be negative")
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	drafts := false
	if includeDrafts != nil {
		drafts = *includeDrafts
	}

	offset := int64(page)
	if page > math.MaxInt64/pageSize {
		offset = math.MaxInt64
	} else {
		offset *= int64(pageSize)
	}

	pages, totalSize, err := s.pages.ListPages(ctx, namespace, drafts, timex.Now(), int32(pageSize), offset)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if int64(pageSize) < int64(totalSize)-offset {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &PageList{Pages: pages, TotalSize: totalSize, NextPageToken: nextPageToken}, nil
}

func (s *Service) UpdatePage(ctx context.Context, parameters UpdatePageParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	item, err := s.pages.FindPageByID(ctx, parameters.Namespace, parameters.ID)
	if err != nil {
		return err
	}

	contentChanged := item.Title != parameters.Title || item.HTML != parameters.HTML
	item.Slug = parameters.Slug
	item.Title = parameters.Title
	item.HTML = parameters.HTML
	item.PublishedAt = parameters.PublishedAt
	now := timex.Now()
	item.UpdatedAt = &now

	if !contentChanged {
		return s.pages.UpdatePage(ctx, item, nil)
	}

	contentID := uuid.New()
	if err := s.pages.UpdatePage(ctx, item, &contentID); err != nil {
		return err
	}

	return s.pages.CreatePageContent(ctx, item.ID, contentID, item.Title, item.HTML, *item.UpdatedAt)
}

func (s *Service) GetPageVersion(ctx context.Context, namespace string, pageID, contentID uuid.UUID) (*PageVersion, error) {
	return s.pages.GetPageVersion(ctx, namespace, pageID, contentID)
}

func (s *Service) ListPageVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PageVersion, error) {
	return s.pages.ListPageVersions(ctx, namespace, id)
}

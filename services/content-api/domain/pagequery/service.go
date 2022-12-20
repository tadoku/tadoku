package pagequery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrPageNotFound = errors.New("page not found")
var ErrRequestInvalid = errors.New("request is invalid")
var ErrForbidden = errors.New("not allowed")

type PageRepository interface {
	FindBySlug(context.Context, *PageFindRequest) (*PageFindResponse, error)
	ListPages(context.Context, *PageListRequest) (*PageListResponse, error)
}

type Service interface {
	FindBySlug(context.Context, *PageFindRequest) (*PageFindResponse, error)
	ListPages(context.Context, *PageListRequest) (*PageListResponse, error)
}

type service struct {
	pr       PageRepository
	validate *validator.Validate
}

func NewService(pr PageRepository) Service {
	return &service{
		pr:       pr,
		validate: validator.New(),
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
	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
	}

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
	if !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrForbidden
	}

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestInvalid, err)
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.pr.ListPages(ctx, req)
}

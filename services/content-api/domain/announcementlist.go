package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type AnnouncementListRepository interface {
	ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*AnnouncementListResult, error)
}

type AnnouncementListResult struct {
	Announcements []Announcement
	TotalSize     int
	NextPageToken string
}

type AnnouncementListRequest struct {
	Namespace string `validate:"required"`
	PageSize  int
	Page      int
}

type AnnouncementListResponse struct {
	Announcements []Announcement
	TotalSize     int
	NextPageToken string
}

type AnnouncementList struct {
	repo     AnnouncementListRepository
	validate *validator.Validate
}

func NewAnnouncementList(repo AnnouncementListRepository) *AnnouncementList {
	return &AnnouncementList{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *AnnouncementList) Execute(ctx context.Context, req *AnnouncementListRequest) (*AnnouncementListResponse, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestInvalid, err)
	}

	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := s.repo.ListAnnouncements(ctx, req.Namespace, pageSize, req.Page)
	if err != nil {
		return nil, err
	}

	return &AnnouncementListResponse{
		Announcements: result.Announcements,
		TotalSize:     result.TotalSize,
		NextPageToken: result.NextPageToken,
	}, nil
}

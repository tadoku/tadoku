package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogListForUserRepository interface {
	ListLogsForUser(context.Context, *LogListForUserRequest) (*LogListForUserResponse, error)
}

type LogListForUserRequest struct {
	UserID         uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type LogListForUserResponse struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

type LogListForUser struct {
	repo LogListForUserRepository
}

func NewLogListForUser(repo LogListForUserRepository) *LogListForUser {
	return &LogListForUser{repo: repo}
}

func (s *LogListForUser) Execute(ctx context.Context, req *LogListForUserRequest) (*LogListForUserResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	if req.PageSize > 100 || req.PageSize < 0 {
		req.PageSize = 100
	}

	if req.IncludeDeleted && !commondomain.IsRole(ctx, commondomain.RoleAdmin) {
		return nil, ErrUnauthorized
	}

	return s.repo.ListLogsForUser(ctx, req)
}

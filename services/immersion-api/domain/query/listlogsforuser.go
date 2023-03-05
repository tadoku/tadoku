package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type ListLogsForUserRequest struct {
	UserID         uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type ListLogsForUserResponse struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

func (s *ServiceImpl) ListLogsForUser(ctx context.Context, req *ListLogsForUserRequest) (*ListLogsForUserResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	if req.PageSize > 100 || req.PageSize < 0 {
		req.PageSize = 100
	}

	if req.IncludeDeleted && !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrUnauthorized
	}

	return s.r.ListLogsForUser(ctx, req)
}

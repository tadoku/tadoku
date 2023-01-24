package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type ListLogsForContestUserRequest struct {
	UserID         uuid.UUID
	ContestID      uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type ListLogsForContestUserResponse struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

func (s *ServiceImpl) ListLogsForContestUser(ctx context.Context, req *ListLogsForContestUserRequest) (*ListLogsForContestUserResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	if req.PageSize > 100 || req.PageSize < 0 {
		req.PageSize = 100
	}

	if req.IncludeDeleted && !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrUnauthorized
	}

	return s.r.ListLogsForContestUser(ctx, req)
}
package domain

import (
	"context"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type ContestListRepository interface {
	ListContests(context.Context, *ContestListRequest) (*ContestListResponse, error)
}

type ContestListRequest struct {
	UserID         uuid.NullUUID
	OfficialOnly   bool
	IncludeDeleted bool
	IncludePrivate bool
	PageSize       int
	Page           int
}

type ContestListResponse struct {
	Contests      []Contest
	TotalSize     int
	NextPageToken string
}

type ContestList struct {
	repo ContestListRepository
}

func NewContestList(repo ContestListRepository) *ContestList {
	return &ContestList{repo: repo}
}

func (s *ContestList) Execute(ctx context.Context, req *ContestListRequest) (*ContestListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 || req.PageSize == 0 {
		req.PageSize = 100
	}

	req.IncludePrivate = commondomain.IsRole(ctx, commondomain.RoleAdmin)

	return s.repo.ListContests(ctx, req)
}

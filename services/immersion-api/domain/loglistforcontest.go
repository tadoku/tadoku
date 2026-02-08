package domain

import (
	"context"

	"github.com/google/uuid"
)

type LogListForContestRepository interface {
	ListLogsForContest(context.Context, *LogListForContestRequest) (*LogListForContestResponse, error)
}

type LogListForContestRequest struct {
	UserID         uuid.NullUUID
	ContestID      uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type LogListForContestResponse struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

type LogListForContest struct {
	repo LogListForContestRepository
}

func NewLogListForContest(repo LogListForContestRepository) *LogListForContest {
	return &LogListForContest{repo: repo}
}

func (s *LogListForContest) Execute(ctx context.Context, req *LogListForContestRequest) (*LogListForContestResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	if req.PageSize > 100 || req.PageSize < 0 {
		req.PageSize = 100
	}

	if req.IncludeDeleted && !isAdmin(ctx) {
		return nil, ErrUnauthorized
	}

	return s.repo.ListLogsForContest(ctx, req)
}

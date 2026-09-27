package leaderboard

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Request struct {
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

func (r Request) Validate() error {
	if r.PageSize < 0 {
		return errx.NewInvalidInputError("page_size must not be negative")
	}
	if r.Page < 0 {
		return errx.NewInvalidInputError("page must not be negative")
	}
	return validateActivity(r.ActivityID)
}

func (r Request) offset() int32 {
	if r.PageSize > 0 && r.Page > math.MaxInt32/r.PageSize {
		return math.MaxInt32
	}
	return int32(r.Page * r.PageSize)
}

type ContestRequest struct {
	Request
	ContestID uuid.UUID
}

type YearlyRequest struct {
	Request
	Year int32
}

type Leaderboard struct {
	Entries       []Entry
	TotalSize     int
	NextPageToken string
}

type Result struct {
	Leaderboard         *Leaderboard
	HydrateDisplayNames bool
}

type Entry struct {
	Rank            int
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	IsTie           bool
}

type score struct {
	userID uuid.UUID
	value  float64
}

type page struct {
	scores     []score
	totalCount int
	startRank  int
	hasPrevTie bool
	hasNextTie bool
}

func result(entries []Entry, total int, request Request) *Leaderboard {
	next := ""
	if int64(request.offset())+int64(request.PageSize) < int64(total) {
		next = fmt.Sprint(request.Page + 1)
	}
	return &Leaderboard{
		Entries:       entries,
		TotalSize:     total,
		NextPageToken: next,
	}
}

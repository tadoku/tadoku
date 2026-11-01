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

func (r YearlyRequest) Validate() error {
	if err := r.Request.Validate(); err != nil {
		return err
	}
	if r.Year < 1 || r.Year > 9999 {
		return errx.NewInvalidInputError("year must be between 1 and 9999")
	}
	return nil
}

type Leaderboard struct {
	Entries       []Entry
	TotalSize     int
	NextPageToken string
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

// A page holds scores ordered by score descending, then user ID descending, starting at offset.
// aboveFirst counts board entries scoring above the first score, and throughLast counts entries
// scoring at least the last score; both are zero for an empty page.
type page struct {
	scores      []score
	offset      int
	totalCount  int
	aboveFirst  int
	throughLast int
}

func result(board *page, request Request) *Leaderboard {
	next := ""
	if int64(request.offset())+int64(request.PageSize) < int64(board.totalCount) {
		next = fmt.Sprint(request.Page + 1)
	}
	return &Leaderboard{
		Entries:       buildEntries(board),
		TotalSize:     board.totalCount,
		NextPageToken: next,
	}
}

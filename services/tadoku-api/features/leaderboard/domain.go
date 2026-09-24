package leaderboard

import (
	"fmt"

	"github.com/google/uuid"
)

type Request struct {
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
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

type outboxEvent struct {
	id        int64
	eventType string
	contestID *uuid.UUID
	year      *int16
}

type page struct {
	scores     []score
	totalCount int
	startRank  int
	hasPrevTie bool
	hasNextTie bool
}

func result(entries []Entry, total, currentPage, pageSize int) *Leaderboard {
	next := ""
	if currentPage*pageSize+pageSize < total {
		next = fmt.Sprint(currentPage + 1)
	}
	return &Leaderboard{
		Entries:       entries,
		TotalSize:     total,
		NextPageToken: next,
	}
}

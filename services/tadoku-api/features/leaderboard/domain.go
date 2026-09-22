// Package leaderboard owns leaderboard reads and their Valkey cache.
package leaderboard

import "github.com/google/uuid"

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

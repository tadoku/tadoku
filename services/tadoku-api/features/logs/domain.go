// Package logs owns immersion log data, derived statistics and configuration.
package logs

import (
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Unit struct {
	ID            uuid.UUID
	Key           string
	LogActivityID int
	Name          string
	Modifier      float32
	LanguageCode  *string
}

type ConfigurationOptions struct {
	Units                []Unit
	UserLanguageCodes    []string
	ScoringEngineEnabled bool
}

type TagSuggestion struct {
	Tag   string
	Count int
}

type ActivityScore struct {
	Date    time.Time
	Score   float32
	Updates int
}

type YearlyActivity struct {
	Scores       []ActivityScore
	TotalUpdates int
}

type Score struct {
	LanguageCode string
	LanguageName string
	Score        float32
}

type YearlyScores struct {
	Scores       []Score
	OverallScore float32
}

type ActivitySplitScore struct {
	ActivityID   int
	ActivityName string
	Score        float32
}

type ContestActivity struct {
	Date         time.Time
	LanguageCode string
	Score        float32
}

var (
	ErrLogNotFound     = errx.NewNotFoundError("log not found")
	ErrInvalidActivity = errx.NewInvalidInputError("invalid log activity")
)

type Log struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	UserDisplayName *string
	Description     *string
	LanguageCode    string
	LanguageName    string
	Activity        activities.Activity
	UnitID          uuid.UUID
	UnitKey         string
	UnitName        string
	Tags            []string
	Amount          float32
	Modifier        float32
	Score           float32
	DurationSeconds *int32
	CreatedAt       time.Time
	Deleted         bool
	Registrations   []RegistrationReference
}

type RegistrationReference struct {
	RegistrationID       uuid.UUID
	ContestID            uuid.UUID
	ContestEnd           time.Time
	Title                string
	OwnerUserDisplayName string
	Official             bool
	Score                float32
}

type ListParameters struct {
	UserID         *uuid.UUID
	ContestID      uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type LogList struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

func (p ListParameters) normalized() ListParameters {
	if p.PageSize == 0 {
		p.PageSize = 50
	}
	if p.PageSize > 100 || p.PageSize < 0 {
		p.PageSize = 100
	}
	return p
}

func hydrateLogActivity(log *Log) error {
	for _, activity := range activities.All() {
		if activity.ID == log.Activity.ID {
			log.Activity = activity
			return nil
		}
	}
	return ErrInvalidActivity
}

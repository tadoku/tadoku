// Package logs owns immersion log data, derived statistics and configuration.
package logs

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
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
	ErrLogFrozen       = errx.NewConflictError("log is frozen")
)

const descriptionMaxLength = 255

func validateDescription(description *string) error {
	if description != nil && utf8.RuneCountInString(*description) > descriptionMaxLength {
		return errx.NewInvalidInputError("description must be at most 255 characters")
	}
	return nil
}

type Tracking = logscore.Tracking
type ContestTracking = logscore.ContestTracking

type logMutation struct {
	ID                          uuid.UUID
	UserID                      uuid.UUID
	LanguageCode                string
	ActivityID                  int32
	Description                 *string
	Tags                        []string
	Tracking                    Tracking
	ContestTrackings            []ContestTracking
	EligibleOfficialLeaderboard bool
	Year                        int16
	Now                         time.Time
}

type OutboxContext struct {
	UserID           uuid.UUID
	Year             int16
	EligibleOfficial bool
}

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
	Tracking        Tracking
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

// Viewer identifies who reads a log: a guest, a user or an administrator. Only
// the log's owner and administrators see its contest registrations.
type Viewer interface {
	isViewer()
}

type GuestViewer struct{}

func (GuestViewer) isViewer() {}

type UserViewer struct {
	UserID uuid.UUID
}

func (UserViewer) isViewer() {}

type AdminViewer struct{}

func (AdminViewer) isViewer() {}

func mayViewRegistrations(viewer Viewer, ownerID uuid.UUID) bool {
	switch viewer := viewer.(type) {
	case AdminViewer:
		return true
	case UserViewer:
		return viewer.UserID == ownerID
	default:
		return false
	}
}

type FindForViewerParameters struct {
	Viewer         Viewer
	IncludeDeleted bool
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
	activity, err := findActivity(log.Activity.ID)
	if err != nil {
		return err
	}

	log.Activity = activity
	return nil
}

func findActivity(id int32) (activities.Activity, error) {
	for _, activity := range activities.All() {
		if activity.ID == id {
			return activity, nil
		}
	}
	return activities.Activity{}, ErrInvalidActivity
}

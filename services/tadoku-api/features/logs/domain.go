// Package logs owns immersion log data, derived statistics and configuration.
package logs

import (
	"math"
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
	ErrLogFrozen       = errx.NewConflictError("log is frozen")
)

type Tracking struct {
	UnitID          *uuid.UUID
	UnitKey         string
	Amount          *float32
	Modifier        *float32
	DurationSeconds *int32
	Score           float32
	RuleSetID       *uuid.UUID
	RuleIDs         []uuid.UUID
	Rates           []float32
	Source          string
}

type ContestTracking struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Tracking       Tracking
}

type Mutation struct {
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

func ValidateAndResolveTracking(activityID int32, unit *Unit, unitID *uuid.UUID, unitKey *string, amount *float32, duration *int32) (Tracking, error) {
	if activityID < 1 || activityID > 5 {
		return Tracking{}, errx.NewInvalidInputError("activity_id is not valid")
	}
	unitActivity, knownUnit := activities.UnitActivityID(stringValue(unitKey))
	if unitKey != nil && (!knownUnit || unitActivity != activityID) {
		return Tracking{}, errx.NewInvalidInputError("unit_key is not valid for activity_id")
	}
	if unit != nil {
		unitActivity, knownUnit = activities.UnitActivityID(unit.Key)
	}
	if unit != nil && (!knownUnit || unitActivity != activityID) {
		return Tracking{}, errx.NewInvalidInputError("resolved unit is not valid for activity_id")
	}
	if unit != nil && unitKey != nil && unit.Key != *unitKey {
		return Tracking{}, errx.NewInvalidInputError("unit_id and unit_key identify different units")
	}
	hasAmount := amount != nil
	hasUnit := unitID != nil || unitKey != nil
	if duration != nil && *duration <= 0 {
		return Tracking{}, errx.NewInvalidInputError("duration_seconds must be positive")
	}
	if amount != nil && (!finite(*amount) || *amount <= 0) {
		return Tracking{}, errx.NewInvalidInputError("amount must be positive")
	}
	if hasAmount != hasUnit {
		return Tracking{}, errx.NewInvalidInputError("amount and a unit identifier must be supplied together")
	}
	if !hasAmount && duration == nil {
		return Tracking{}, errx.NewInvalidInputError("amount/unit or duration_seconds is required")
	}

	tracking := Tracking{DurationSeconds: duration}
	if hasAmount {
		if unit == nil {
			return Tracking{}, errx.NewInvalidInputError("unit is required for amount scoring")
		}
		tracking.UnitID = &unit.ID
		tracking.UnitKey = unit.Key
		tracking.Amount = amount
		tracking.Modifier = &unit.Modifier
		tracking.Score = *amount * unit.Modifier
	} else {
		minutes := float32(*duration) / 60
		rates := [...]float32{0, .2, .4, .2, .5, .5}
		tracking.Score = minutes * rates[activityID]
	}
	return tracking, nil
}

func finite(value float32) bool { return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) }
func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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

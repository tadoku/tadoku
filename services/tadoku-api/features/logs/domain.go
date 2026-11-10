package logs

import (
	"math"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type ConfigurationOptions struct {
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

type ContestProfileSummary struct {
	Score      float32
	LogCount   int
	ActiveDays int
	BestDay    *DayScore
	Days       []ContestDayScore
	Languages  []Score
	Activities []ActivitySplitScore
	Tags       []TagScore
}

type DayScore struct {
	Date  time.Time
	Score float32
}

type ContestDayScore struct {
	Date       time.Time
	ActivityID int
	Score      float32
	logCount   int
}

type TagScore struct {
	Tag      string
	Score    float32
	LogCount int
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

func validateYear(year int) error {
	if year < 1 || year > 9999 {
		return errx.NewInvalidInputError("year must be between 1 and 9999")
	}
	return nil
}

type lockedLog struct {
	UserID uuid.UUID
	Frozen bool
}

type logMutation struct {
	ID                          uuid.UUID
	UserID                      uuid.UUID
	LanguageCode                string
	ActivityID                  int32
	Description                 *string
	Tags                        []string
	Tracking                    logscore.Tracking
	ContestTrackings            []logscore.ContestTracking
	EligibleOfficialLeaderboard bool
	Year                        int16
	Now                         time.Time
}

type Log struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	UserDisplayName *string
	Description     *string
	LanguageCode    string
	LanguageName    string
	Activity        activities.Activity
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
	Tracking        logscore.Tracking
}

type attachedContest struct {
	ID  uuid.UUID
	End time.Time
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
	Filters        Filters
	IncludeDeleted bool
	PageSize       int
	Page           int
}

// Filters selects logs matching any of the values of each non-empty field, and
// every non-empty field.
type Filters struct {
	LanguageCodes []string
	ActivityIDs   []int
	Tags          []string
}

const (
	maxLanguageFilterValues = 10
	maxActivityFilterValues = 10
	maxTagFilterValues      = 20
	maxLanguageCodeLength   = 10
	maxTagLength            = 50
)

func (f Filters) Validate() error {
	if len(f.LanguageCodes) > maxLanguageFilterValues {
		return errx.NewInvalidInputError("language must have at most 10 values")
	}
	for _, code := range f.LanguageCodes {
		length := utf8.RuneCountInString(code)
		if length < 1 || length > maxLanguageCodeLength {
			return errx.NewInvalidInputError("language must be 1 to 10 characters")
		}
	}

	if len(f.ActivityIDs) > maxActivityFilterValues {
		return errx.NewInvalidInputError("activity must have at most 10 values")
	}
	for _, id := range f.ActivityIDs {
		if id < 1 || id > math.MaxInt16 {
			return errx.NewInvalidInputError("activity must be between 1 and 32767")
		}
	}

	if len(f.Tags) > maxTagFilterValues {
		return errx.NewInvalidInputError("tag must have at most 20 values")
	}
	for _, tag := range f.Tags {
		length := utf8.RuneCountInString(tag)
		if length < 1 || length > maxTagLength {
			return errx.NewInvalidInputError("tag must be 1 to 50 characters")
		}
	}
	return nil
}

type queryFilters struct {
	languageCodes []string
	activityIDs   []int16
	tags          []string
}

// query must follow a successful Validate. An empty field becomes nil, which
// the queries read as no filter.
func (f Filters) query() queryFilters {
	result := queryFilters{}
	if len(f.LanguageCodes) > 0 {
		result.languageCodes = f.LanguageCodes
	}
	if len(f.Tags) > 0 {
		result.tags = f.Tags
	}
	for _, id := range f.ActivityIDs {
		result.activityIDs = append(result.activityIDs, int16(id))
	}
	return result
}

type LogList struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

func (p ListParameters) Validate() error {
	if err := p.Filters.Validate(); err != nil {
		return err
	}
	if p.PageSize < 0 {
		return errx.NewInvalidInputError("page_size must not be negative")
	}
	if p.Page < 0 {
		return errx.NewInvalidInputError("page must not be negative")
	}
	return nil
}

func (p ListParameters) normalized() ListParameters {
	if p.PageSize == 0 {
		p.PageSize = 50
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p
}

func (p ListParameters) offset() int32 {
	if p.PageSize > 0 && p.Page > math.MaxInt32/p.PageSize {
		return math.MaxInt32
	}
	return int32(p.Page * p.PageSize)
}

func unitName(activityID int32, key, languageCode string) string {
	unit, err := activities.ResolveUnit(activityID, key, languageCode)
	if err != nil {
		return ""
	}
	return unit.Name
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

type CreateResult struct {
	ID     uuid.UUID
	Events []contestevent.Change
}

type MutationResult struct {
	Events []contestevent.Change
}

type RemovalResult struct {
	Events      []contestevent.Change
	DeletedLogs int
	KeptLogs    int
}

type ExportedLog struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	LanguageCode    string
	LanguageName    string
	Activity        activities.Activity
	Amount          *ExportedAmount
	DurationSeconds *int32
	Score           float32
	Tags            []string
	Description     *string
}

type ExportedAmount struct {
	Value    float32
	UnitName string
}

type exportCursor struct {
	createdAt time.Time
	id        uuid.UUID
}

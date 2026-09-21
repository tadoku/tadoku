// Package contests owns contest discovery and persistence.
package contests

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type ActivityInputType string

const (
	ActivityInputTypeAmountPrimary ActivityInputType = "amount_primary"
	ActivityInputTypeTimePrimary   ActivityInputType = "time_primary"
)

var activities = []Activity{
	{ID: 1, Name: "Reading", Default: true, InputType: ActivityInputTypeAmountPrimary},
	{ID: 2, Name: "Listening", Default: true, InputType: ActivityInputTypeTimePrimary},
	{ID: 3, Name: "Writing", Default: false, InputType: ActivityInputTypeAmountPrimary},
	{ID: 4, Name: "Speaking", Default: false, InputType: ActivityInputTypeTimePrimary},
	{ID: 5, Name: "Study", Default: false, InputType: ActivityInputTypeTimePrimary},
}

var (
	ErrContestNotFound = errx.NewNotFoundError("contest not found")
	ErrInvalidActivity = errx.NewInvalidInputError("invalid contest activity")
)

type Language struct {
	Code string
	Name string
}

type Activity struct {
	ID        int32
	Name      string
	Default   bool
	InputType ActivityInputType
}

type Contest struct {
	ID                      uuid.UUID
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationEnd         time.Time
	Title                   string
	Description             *string
	OwnerUserID             uuid.UUID
	OwnerUserDisplayName    string
	Official                bool
	Private                 bool
	LanguageCodeAllowList   []string
	ActivityTypeIDAllowList []int32
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Deleted                 bool
}

type ContestView struct {
	ID                   uuid.UUID
	ContestStart         time.Time
	ContestEnd           time.Time
	RegistrationEnd      time.Time
	Title                string
	Description          *string
	OwnerUserID          uuid.UUID
	OwnerUserDisplayName string
	Official             bool
	Private              bool
	AllowedLanguages     []Language
	AllowedActivities    []Activity
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Deleted              bool
}

type ContestList struct {
	Contests      []Contest
	TotalSize     int
	NextPageToken string
}

type ConfigurationOptions struct {
	Languages              []Language
	Activities             []Activity
	CanCreateOfficialRound bool
}

type ListParameters struct {
	UserID         *uuid.UUID
	Official       bool
	IncludeDeleted bool
	PageSize       int
	Page           int

	includePrivate bool
}

func (p ListParameters) IncludePrivate() bool { return p.includePrivate }

type FindParameters struct {
	ID uuid.UUID

	includeDeleted bool
}

func (p FindParameters) IncludeDeleted() bool { return p.includeDeleted }

func allActivities() []Activity {
	return append([]Activity(nil), activities...)
}

func hydrateActivities(ids []int32) ([]Activity, error) {
	result := make([]Activity, 0, len(ids))
	for _, id := range ids {
		if id < 1 || int(id) > len(activities) || activities[id-1].ID != id {
			return nil, ErrInvalidActivity
		}
		result = append(result, activities[id-1])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

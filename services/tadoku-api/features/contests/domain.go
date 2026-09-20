// Package contests owns contest discovery and persistence.
package contests

import (
	"errors"
	"sort"
	"time"
	"unicode/utf8"

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
	ErrContestNotFound           = errx.NewNotFoundError("contest not found")
	ErrContestCreatorNotFound    = errx.NewNotFoundError("contest creator not found")
	ErrContestCreationForbidden  = errx.NewForbiddenError("contest creation forbidden")
	ErrInvalidContestCreator     = errors.New("invalid contest creator identity")
	ErrContestCreatorTooYoung    = errors.New("contest creator account too young")
	ErrInvalidActivity           = errx.NewInvalidInputError("invalid contest activity")
	ErrInvalidContest            = errx.NewInvalidInputError("invalid contest")
	ErrAccountDeletionInProgress = errx.NewConflictError("account deletion in progress")
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

type CreateParameters struct {
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationEnd         time.Time
	Title                   string
	Description             *string
	Official                bool
	Private                 bool
	LanguageCodeAllowList   []string
	ActivityTypeIDAllowList []int32

	id                   uuid.UUID
	ownerUserID          uuid.UUID
	ownerUserDisplayName string
	sessionCreatedAt     time.Time
	createdAt            time.Time
	updatedAt            time.Time
}

func (p CreateParameters) ID() uuid.UUID                { return p.id }
func (p CreateParameters) OwnerUserID() uuid.UUID       { return p.ownerUserID }
func (p CreateParameters) OwnerUserDisplayName() string { return p.ownerUserDisplayName }
func (p CreateParameters) SessionCreatedAt() time.Time  { return p.sessionCreatedAt }
func (p CreateParameters) CreatedAt() time.Time         { return p.createdAt }
func (p CreateParameters) UpdatedAt() time.Time         { return p.updatedAt }

func (p CreateParameters) validate(admin bool, now time.Time) error {
	if p.ownerUserID == uuid.Nil || p.ownerUserDisplayName == "" ||
		p.ContestStart.IsZero() || p.ContestEnd.IsZero() || p.RegistrationEnd.IsZero() ||
		utf8.RuneCountInString(p.Title) <= 3 || len(p.ActivityTypeIDAllowList) == 0 {
		return ErrInvalidContest
	}
	if p.Official && (p.Private || len(p.LanguageCodeAllowList) != 0) {
		return ErrInvalidContest
	}
	if p.ContestStart.After(p.ContestEnd) {
		return ErrInvalidContest
	}
	for _, id := range p.ActivityTypeIDAllowList {
		if id < 1 || int(id) > len(activities) || activities[id-1].ID != id {
			return ErrInvalidContest
		}
	}
	if !admin {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if p.ContestStart.Before(today) || p.ContestEnd.Before(today) {
			return ErrInvalidContest
		}
	}
	return nil
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

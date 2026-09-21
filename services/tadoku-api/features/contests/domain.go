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
	ErrContestNotFound          = errx.NewNotFoundError("contest not found")
	ErrContestCreatorNotFound   = errx.NewNotFoundError("contest creator not found")
	ErrContestCreationForbidden = errx.NewForbiddenError("contest creation forbidden")
	ErrContestCreatorTooYoung   = errors.New("contest creator account too young")
	ErrInvalidActivity          = errx.NewInvalidInputError("invalid contest activity")
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

type CreateContestParameters struct {
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
	createdAt            time.Time
	updatedAt            time.Time
}

func (p CreateContestParameters) ID() uuid.UUID                { return p.id }
func (p CreateContestParameters) OwnerUserID() uuid.UUID       { return p.ownerUserID }
func (p CreateContestParameters) OwnerUserDisplayName() string { return p.ownerUserDisplayName }
func (p CreateContestParameters) CreatedAt() time.Time         { return p.createdAt }
func (p CreateContestParameters) UpdatedAt() time.Time         { return p.updatedAt }

func (p CreateContestParameters) validate(admin bool, now time.Time) error {
	if p.ownerUserID == uuid.Nil {
		return errx.NewInvalidInputError("invalid contest OwnerUserID: must not be nil")
	}
	if p.ownerUserDisplayName == "" {
		return errx.NewInvalidInputError("invalid contest OwnerUserDisplayName: must not be empty")
	}
	if p.ContestStart.IsZero() {
		return errx.NewInvalidInputError("invalid contest ContestStart: must not be zero")
	}
	if p.ContestEnd.IsZero() {
		return errx.NewInvalidInputError("invalid contest ContestEnd: must not be zero")
	}
	if p.RegistrationEnd.IsZero() {
		return errx.NewInvalidInputError("invalid contest RegistrationEnd: must not be zero")
	}
	if utf8.RuneCountInString(p.Title) <= 3 {
		return errx.NewInvalidInputError("invalid contest Title: must contain more than three characters")
	}
	if len(p.ActivityTypeIDAllowList) == 0 {
		return errx.NewInvalidInputError("invalid contest ActivityTypeIDAllowList: must not be empty")
	}
	if p.Official && p.Private {
		return errx.NewInvalidInputError("invalid contest Private: official contests must be public")
	}
	if p.Official && len(p.LanguageCodeAllowList) != 0 {
		return errx.NewInvalidInputError("invalid contest LanguageCodeAllowList: official contests must allow every language")
	}
	if p.ContestStart.After(p.ContestEnd) {
		return errx.NewInvalidInputError("invalid contest ContestStart: must not be after ContestEnd")
	}
	for _, id := range p.ActivityTypeIDAllowList {
		if id < 1 || int(id) > len(activities) || activities[id-1].ID != id {
			return errx.NewInvalidInputError("invalid contest ActivityTypeIDAllowList: contains an unknown activity")
		}
	}
	if !admin {
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if p.ContestStart.Before(today) {
			return errx.NewInvalidInputError("invalid contest ContestStart: non-admin contests must not start in the past")
		}
		if p.ContestEnd.Before(today) {
			return errx.NewInvalidInputError("invalid contest ContestEnd: non-admin contests must not end in the past")
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

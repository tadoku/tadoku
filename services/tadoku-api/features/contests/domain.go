// Package contests owns contest discovery and persistence.
package contests

import (
	"errors"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	activitiescatalog "github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var activities = activitiescatalog.All()

var (
	ErrContestNotFound          = errx.NewNotFoundError("contest not found")
	ErrContestCreatorNotFound   = errx.NewNotFoundError("contest creator not found")
	ErrContestCreationForbidden = errx.NewForbiddenError("contest creation forbidden")
	ErrContestCreatorTooYoung   = errors.New("contest creator account too young")
	ErrInvalidActivity          = errx.NewInvalidInputError("invalid contest activity")
	ErrInvalidRegistration      = errx.NewInvalidInputError("invalid contest registration")
	ErrRegistrationNotFound     = errx.NewNotFoundError("contest registration not found")
)

type Language struct {
	Code string
	Name string
}

type Activity = activitiescatalog.Activity

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
	allowedActivityIDs   []int32
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Deleted              bool
}

type ContestList struct {
	Contests      []Contest
	TotalSize     int
	NextPageToken string
}

type ContestSummary struct {
	ParticipantCount int
	LanguageCount    int
	TotalScore       float32
}

type Registration struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Languages []Language
	Contest   *ContestView
}

func (r Registration) IsEligibleForScoring(languageCode string, activityID int32) bool {
	languageAllowed := false
	for _, code := range r.LanguageCodes {
		if code == languageCode {
			languageAllowed = true
			break
		}
	}
	if !languageAllowed || r.Contest == nil {
		return false
	}

	for _, activity := range r.Contest.AllowedActivities {
		if activity.ID == activityID {
			return true
		}
	}
	return false
}

func selectRegistrationsForScoring(requested []uuid.UUID, available []Registration, languageCode string, activityID int32) ([]Registration, error) {
	byID := make(map[uuid.UUID]Registration, len(available))
	for _, registration := range available {
		byID[registration.ID] = registration
	}

	selected := make([]Registration, 0, len(requested))
	for _, id := range requested {
		registration, exists := byID[id]
		if !exists {
			return nil, errx.NewInvalidInputError("registration_id is not ongoing for the current user")
		}
		if !registration.IsEligibleForScoring(languageCode, activityID) {
			return nil, errx.NewInvalidInputError("language_code or activity_id is not allowed for registration_id")
		}
		selected = append(selected, registration)
	}

	return selected, nil
}

type RegistrationList struct {
	Registrations []Registration
	TotalSize     int
	NextPageToken string
}

type RegistrationUpsertParameters struct {
	ContestID     uuid.UUID
	LanguageCodes []string
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
}

func (p CreateContestParameters) validate(ownerUserID uuid.UUID, ownerUserDisplayName string, admin bool, now time.Time) error {
	if ownerUserID == uuid.Nil {
		return errx.NewInvalidInputError("invalid contest OwnerUserID: must not be nil")
	}
	if ownerUserDisplayName == "" {
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
	result, err := hydrateActivitiesInOrder(ids)
	if err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func hydrateActivitiesInOrder(ids []int32) ([]Activity, error) {
	result := make([]Activity, 0, len(ids))
	for _, id := range ids {
		if id < 1 || int(id) > len(activities) || activities[id-1].ID != id {
			return nil, ErrInvalidActivity
		}
		result = append(result, activities[id-1])
	}
	return result, nil
}

package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

const (
	MaxTagsPerLog = 10
	MaxTagLength  = 50
)

// ValidateAndNormalizeTags validates and normalizes tags:
// - Trims whitespace and converts to lowercase
// - Removes empty and duplicate tags
// - Enforces max 10 tags, 50 chars each
func ValidateAndNormalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" || len(normalized) > MaxTagLength || seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	if len(result) > MaxTagsPerLog {
		return nil, fmt.Errorf("maximum %d tags allowed", MaxTagsPerLog)
	}
	return result, nil
}

type CreateLogRequest struct {
	RegistrationIDs []uuid.UUID `validate:"required"`
	UnitID          uuid.UUID   `validate:"required"`
	UserID          uuid.UUID   `validate:"required"`
	ActivityID      int32       `validate:"required"`
	LanguageCode    string      `validate:"required"`
	Amount          float32     `validate:"required,gte=0"`
	Tags            []string

	// Optional
	Description                 *string
	EligibleOfficialLeaderboard bool
}

func (s *ServiceImpl) CreateLog(ctx context.Context, req *CreateLogRequest) (*query.Log, error) {
	// Make sure the user is authorized to create a contest
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	if err := s.UpdateUserMetadataFromSession(ctx); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	req.UserID = uuid.MustParse(session.Subject)

	err := s.validate.Struct(req)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

	// Validate and normalize tags
	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags: %w", ErrInvalidLog)
	}

	registrations, err := s.r.FetchOngoingContestRegistrations(ctx, &query.FetchOngoingContestRegistrationsRequest{
		UserID: req.UserID,
		Now:    s.clock.Now(),
	})
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("unable to fetch registrations: %w", err)
	}

	validContestIDs := map[uuid.UUID]query.ContestRegistration{}
	for _, r := range registrations.Registrations {
		validContestIDs[r.ID] = r
	}

	// validate registrations
	for _, id := range req.RegistrationIDs {
		registration, ok := validContestIDs[id]
		if !ok {
			return nil, fmt.Errorf("registration is not found as ongoing for the current user: %w", ErrInvalidLog)
		}

		if registration.Contest.Official {
			req.EligibleOfficialLeaderboard = true
		}

		// validate language is part of registration
		found := false
		for _, lang := range registration.Languages {
			if lang.Code == req.LanguageCode {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("language is not allowed for registration: %w", ErrInvalidLog)
		}

		// validate activity is allowed by the contest
		found = false
		for _, act := range registration.Contest.AllowedActivities {
			if act.ID == req.ActivityID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("activity is not allowed for registration: %w", ErrInvalidLog)
		}
	}

	logId, err := s.r.CreateLog(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	return s.r.FindLogByID(ctx, &query.FindLogByIDRequest{
		ID:             *logId,
		IncludeDeleted: false,
	})
}

package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogCreateRepository interface {
	FetchOngoingContestRegistrations(context.Context, *RegistrationListOngoingRequest) (*ContestRegistrations, error)
	CreateLog(context.Context, *LogCreateRequest) (*uuid.UUID, error)
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	FindActivityByID(context.Context, int32) (*Activity, error)
}

type LogCreateRequest struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	ActivityID      int32  `validate:"required"`
	LanguageCode    string `validate:"required"`
	Amount          *float32
	DurationSeconds *int32
	Tags            []string

	// Optional
	Description *string

	// Set by domain layer (unexported: only domain can write, others read via getters)
	userID                      uuid.UUID
	eligibleOfficialLeaderboard bool
	year                        int16
	activity                    *Activity
}

func (r *LogCreateRequest) UserID() uuid.UUID                 { return r.userID }
func (r *LogCreateRequest) EligibleOfficialLeaderboard() bool { return r.eligibleOfficialLeaderboard }
func (r *LogCreateRequest) Year() int16                       { return r.year }
func (r *LogCreateRequest) Activity() *Activity               { return r.activity }

type LogCreate struct {
	repo       LogCreateRepository
	clock      commondomain.Clock
	validate   *validator.Validate
	userUpsert *UserUpsert
}

func NewLogCreate(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
) *LogCreate {
	return &LogCreate{
		repo:       repo,
		clock:      clock,
		validate:   validator.New(),
		userUpsert: userUpsert,
	}
}

func (s *LogCreate) Execute(ctx context.Context, req *LogCreateRequest) (*Log, error) {
	// Make sure the user is authorized to create a log
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}

	if err := s.userUpsert.Execute(ctx); err != nil {
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	// Enrich request with session
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	req.userID = uuid.MustParse(session.Subject)

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

	// Look up activity to determine input type
	activity, err := s.repo.FindActivityByID(ctx, req.ActivityID)
	if err != nil {
		return nil, fmt.Errorf("could not find activity: %w", ErrInvalidLog)
	}

	// Validate tracking data based on input type
	if err := validateTrackingData(activity, req.DurationSeconds, req.Amount, req.UnitID); err != nil {
		return nil, err
	}

	// Store activity for repo layer to use when computing score
	req.activity = activity

	// Validate and normalize tags
	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	if len(req.RegistrationIDs) > 0 {
		registrations, fetchErr := s.repo.FetchOngoingContestRegistrations(ctx, &RegistrationListOngoingRequest{
			UserID: req.userID,
			Now:    s.clock.Now(),
		})
		if fetchErr != nil {
			return nil, fmt.Errorf("unable to fetch registrations: %w", fetchErr)
		}

		validContestIDs := map[uuid.UUID]ContestRegistration{}
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
				req.eligibleOfficialLeaderboard = true
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
	}

	req.year = int16(s.clock.Now().Year())

	logId, err := s.repo.CreateLog(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not create log: %w", err)
	}

	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             *logId,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, err
	}

	return log, nil
}

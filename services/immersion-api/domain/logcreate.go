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
}

type LogCreateRequest struct {
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

type LogCreate struct {
	repo               LogCreateRepository
	clock              commondomain.Clock
	validate           *validator.Validate
	userUpsert         *UserUpsert
	leaderboardUpdater LeaderboardScoreUpdater
}

func NewLogCreate(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
	leaderboardUpdater LeaderboardScoreUpdater,
) *LogCreate {
	return &LogCreate{
		repo:               repo,
		clock:              clock,
		validate:           validator.New(),
		userUpsert:         userUpsert,
		leaderboardUpdater: leaderboardUpdater,
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
	req.UserID = uuid.MustParse(session.Subject)

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

	// Validate and normalize tags
	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	registrations, err := s.repo.FetchOngoingContestRegistrations(ctx, &RegistrationListOngoingRequest{
		UserID: req.UserID,
		Now:    s.clock.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch registrations: %w", err)
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

	// Update leaderboards — best effort, do not fail the log creation
	s.updateLeaderboards(ctx, req, validContestIDs, log)

	return log, nil
}

// updateLeaderboards recalculates the user's score for all relevant leaderboards.
// For each attached contest: recalculate the user's total contest score.
// For official contests: also recalculate yearly and global scores (pipelined).
func (s *LogCreate) updateLeaderboards(ctx context.Context, req *LogCreateRequest, validContestIDs map[uuid.UUID]ContestRegistration, log *Log) {
	year := log.CreatedAt.Year()

	for _, regID := range req.RegistrationIDs {
		registration := validContestIDs[regID]
		s.leaderboardUpdater.UpdateUserContestScore(ctx, registration.ContestID, req.UserID)
	}

	if req.EligibleOfficialLeaderboard {
		s.leaderboardUpdater.UpdateUserOfficialScores(ctx, year, req.UserID)
	}
}

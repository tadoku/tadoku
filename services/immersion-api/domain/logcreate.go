package domain

import (
	"context"
	"fmt"
	"log/slog"

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
	repo             LogCreateRepository
	clock            commondomain.Clock
	validate         *validator.Validate
	userUpsert       *UserUpsert
	leaderboardStore LeaderboardStore
	leaderboardRepo  LeaderboardRebuildRepository
}

func NewLogCreate(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
	leaderboardStore LeaderboardStore,
	leaderboardRepo LeaderboardRebuildRepository,
) *LogCreate {
	return &LogCreate{
		repo:             repo,
		clock:            clock,
		validate:         validator.New(),
		userUpsert:       userUpsert,
		leaderboardStore: leaderboardStore,
		leaderboardRepo:  leaderboardRepo,
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

	// Update leaderboards in Redis — best effort, do not fail the log creation
	s.updateLeaderboards(ctx, req, validContestIDs, log)

	return log, nil
}

// updateLeaderboards updates all relevant Redis leaderboards after a log is created.
// For each attached contest: increment the user's score (or rebuild if the leaderboard doesn't exist).
// For official contests: also update yearly and global leaderboards.
// Errors are logged but do not fail the log creation.
func (s *LogCreate) updateLeaderboards(ctx context.Context, req *LogCreateRequest, validContestIDs map[uuid.UUID]ContestRegistration, log *Log) {
	score := float64(log.Score)
	year := log.CreatedAt.Year()

	for _, regID := range req.RegistrationIDs {
		registration := validContestIDs[regID]
		contestID := registration.ContestID

		s.updateContestLeaderboard(ctx, contestID, req.UserID, score)
	}

	if req.EligibleOfficialLeaderboard {
		s.updateYearlyLeaderboard(ctx, year, req.UserID, score)
		s.updateGlobalLeaderboard(ctx, req.UserID, score)
	}
}

func (s *LogCreate) updateContestLeaderboard(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) {
	updated, err := s.leaderboardStore.IncrementContestScore(ctx, contestID, userID, score)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment contest leaderboard", "contest_id", contestID, "error", err)
		return
	}
	if updated {
		return
	}

	// Leaderboard doesn't exist in Redis yet, rebuild from database
	scores, err := s.leaderboardRepo.FetchAllContestLeaderboardScores(ctx, contestID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch contest leaderboard scores for rebuild", "contest_id", contestID, "error", err)
		return
	}

	if err := s.leaderboardStore.RebuildContestLeaderboard(ctx, contestID, scores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild contest leaderboard", "contest_id", contestID, "error", err)
	}
}

func (s *LogCreate) updateYearlyLeaderboard(ctx context.Context, year int, userID uuid.UUID, score float64) {
	updated, err := s.leaderboardStore.IncrementYearlyScore(ctx, year, userID, score)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment yearly leaderboard", "year", year, "error", err)
		return
	}
	if updated {
		return
	}

	scores, err := s.leaderboardRepo.FetchAllYearlyLeaderboardScores(ctx, year)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch yearly leaderboard scores for rebuild", "year", year, "error", err)
		return
	}

	if err := s.leaderboardStore.RebuildYearlyLeaderboard(ctx, year, scores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild yearly leaderboard", "year", year, "error", err)
	}
}

func (s *LogCreate) updateGlobalLeaderboard(ctx context.Context, userID uuid.UUID, score float64) {
	updated, err := s.leaderboardStore.IncrementGlobalScore(ctx, userID, score)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment global leaderboard", "error", err)
		return
	}
	if updated {
		return
	}

	scores, err := s.leaderboardRepo.FetchAllGlobalLeaderboardScores(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch global leaderboard scores for rebuild", "error", err)
		return
	}

	if err := s.leaderboardStore.RebuildGlobalLeaderboard(ctx, scores); err != nil {
		slog.ErrorContext(ctx, "failed to rebuild global leaderboard", "error", err)
	}
}

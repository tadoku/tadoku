package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type LogUpdateRepository interface {
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
	FindUnitForTracking(context.Context, *UnitFindForTrackingRequest) (*Unit, error)
	FindUnitForTrackingByKey(context.Context, *UnitFindForTrackingByKeyRequest) (*Unit, error)
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
	FindContestScoringRuleSets(context.Context, uuid.UUID) (*ScoringRuleSet, *ScoringRuleSet, error)
	UpdateLog(context.Context, *LogUpdateRequest) error
}

type LogUpdateRequest struct {
	LogID           uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string

	// Set by domain layer (unexported: only domain can write, others read via getters)
	now              time.Time
	userID           uuid.UUID
	tracking         LogTracking
	contestTrackings []ContestLogTracking
}

func (r *LogUpdateRequest) Now() time.Time        { return r.now }
func (r *LogUpdateRequest) UserID() uuid.UUID     { return r.userID }
func (r *LogUpdateRequest) Tracking() LogTracking { return r.tracking }
func (r *LogUpdateRequest) ContestTrackings() []ContestLogTracking {
	return r.contestTrackings
}

type LogUpdate struct {
	repo             LogUpdateRepository
	clock            commondomain.Clock
	validate         *validator.Validate
	useScoringEngine bool
	scoringObserver  ScoringShadowObserver
}

func NewLogUpdate(
	repo LogUpdateRepository,
	clock commondomain.Clock,
) *LogUpdate {
	return NewLogUpdateWithScoringObserver(repo, clock, false, nil)
}

func NewLogUpdateWithScoringEngine(
	repo LogUpdateRepository,
	clock commondomain.Clock,
	enabled bool,
) *LogUpdate {
	return NewLogUpdateWithScoringObserver(repo, clock, enabled, nil)
}

func NewLogUpdateWithScoringObserver(
	repo LogUpdateRepository,
	clock commondomain.Clock,
	enabled bool,
	observer ScoringShadowObserver,
) *LogUpdate {
	return &LogUpdate{
		repo:             repo,
		clock:            clock,
		validate:         validator.New(),
		useScoringEngine: enabled,
		scoringObserver:  observer,
	}
}

func (s *LogUpdate) Execute(ctx context.Context, req *LogUpdateRequest) (*Log, error) {
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}

	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	log, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not find log to update: %w", err)
	}

	isOwner := log.UserID == userID
	if !isOwner && !isAdmin(ctx) {
		return nil, ErrForbidden
	}

	err = s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	req.tracking, err = resolveLogTracking(
		ctx,
		s.repo,
		int32(log.ActivityID),
		log.LanguageCode,
		req.UnitID,
		req.UnitKey,
		req.Amount,
		req.DurationSeconds,
	)
	if err != nil {
		return nil, err
	}
	scoringInput := ScoringInput{
		ActivityID:      int32(log.ActivityID),
		UnitKey:         req.tracking.UnitKey,
		LanguageCode:    log.LanguageCode,
		Tags:            req.Tags,
		Amount:          req.Amount,
		DurationSeconds: req.DurationSeconds,
	}
	mode := ScoringShadowModeShadow
	if s.useScoringEngine {
		mode = ScoringShadowModeAuthoritative
	}
	result, scoringErr := EvaluateAndObservePlatformScoring(
		ctx,
		s.repo,
		s.scoringObserver,
		ScoringShadowOperationUpdate,
		mode,
		scoringInput,
		req.tracking.ComputedScore,
	)
	if s.useScoringEngine {
		if scoringErr != nil {
			return nil, fmt.Errorf("could not score log: %w", scoringErr)
		}
		ApplyScoringResult(&req.tracking, result)
		req.now = s.clock.Now()
		for _, registration := range log.Registrations {
			if registration.ContestEnd.Before(req.now) {
				continue
			}
			contestTracking, contestErr := resolveContestLogTracking(
				ctx,
				s.repo,
				registration.ContestID,
				registration.RegistrationID,
				req.tracking,
				scoringInput,
			)
			if contestErr != nil {
				return nil, fmt.Errorf("could not score contest %s: %w", registration.ContestID, contestErr)
			}
			req.contestTrackings = append(req.contestTrackings, contestTracking)
		}
	}
	if !s.useScoringEngine {
		req.now = s.clock.Now()
	}

	req.userID = log.UserID

	if err := s.repo.UpdateLog(ctx, req); err != nil {
		return nil, fmt.Errorf("could not update log: %w", err)
	}

	updated, err := s.repo.FindLogByID(ctx, &LogFindRequest{
		ID:             req.LogID,
		IncludeDeleted: false,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch updated log: %w", err)
	}

	if err := hydrateLogActivity(updated); err != nil {
		return nil, err
	}

	return updated, nil
}

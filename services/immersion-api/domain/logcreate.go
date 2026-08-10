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
	FindUnitForTracking(context.Context, *UnitFindForTrackingRequest) (*Unit, error)
	FindUnitForTrackingByKey(context.Context, *UnitFindForTrackingByKeyRequest) (*Unit, error)
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
	FindContestScoringRuleSets(context.Context, uuid.UUID) (*ScoringRuleSet, *ScoringRuleSet, error)
	CreateLog(context.Context, *LogCreateRequest) (*uuid.UUID, error)
	FindLogByID(context.Context, *LogFindRequest) (*Log, error)
}

type LogCreateRequest struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
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
	tracking                    LogTracking
	contestTrackings            []ContestLogTracking
}

func (r *LogCreateRequest) UserID() uuid.UUID                 { return r.userID }
func (r *LogCreateRequest) EligibleOfficialLeaderboard() bool { return r.eligibleOfficialLeaderboard }
func (r *LogCreateRequest) Year() int16                       { return r.year }
func (r *LogCreateRequest) Tracking() LogTracking             { return r.tracking }
func (r *LogCreateRequest) ContestTrackings() []ContestLogTracking {
	return r.contestTrackings
}

type LogCreate struct {
	repo             LogCreateRepository
	clock            commondomain.Clock
	validate         *validator.Validate
	userUpsert       *UserUpsert
	useScoringEngine bool
	scoringObserver  ScoringShadowObserver
}

func NewLogCreate(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
) *LogCreate {
	return NewLogCreateWithScoringObserver(repo, clock, userUpsert, false, nil)
}

func NewLogCreateWithScoringEngine(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
	enabled bool,
) *LogCreate {
	return NewLogCreateWithScoringObserver(repo, clock, userUpsert, enabled, nil)
}

func NewLogCreateWithScoringObserver(
	repo LogCreateRepository,
	clock commondomain.Clock,
	userUpsert *UserUpsert,
	enabled bool,
	observer ScoringShadowObserver,
) *LogCreate {
	return &LogCreate{
		repo:             repo,
		clock:            clock,
		validate:         validator.New(),
		userUpsert:       userUpsert,
		useScoringEngine: enabled,
		scoringObserver:  observer,
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

	// Validate and normalize tags
	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	validContestRegistrations := map[uuid.UUID]ContestRegistration{}
	if len(req.RegistrationIDs) > 0 {
		registrations, fetchErr := s.repo.FetchOngoingContestRegistrations(ctx, &RegistrationListOngoingRequest{
			UserID: req.userID,
			Now:    s.clock.Now(),
		})
		if fetchErr != nil {
			return nil, fmt.Errorf("unable to fetch registrations: %w", fetchErr)
		}

		for _, r := range registrations.Registrations {
			validContestRegistrations[r.ID] = r
		}

		// validate registrations
		for _, id := range req.RegistrationIDs {
			registration, ok := validContestRegistrations[id]
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

	req.tracking, err = resolveLogTracking(
		ctx,
		s.repo,
		req.ActivityID,
		req.LanguageCode,
		req.UnitID,
		req.UnitKey,
		req.Amount,
		req.DurationSeconds,
	)
	if err != nil {
		return nil, err
	}
	scoringInput := ScoringInput{
		ActivityID:      req.ActivityID,
		UnitKey:         req.tracking.UnitKey,
		LanguageCode:    req.LanguageCode,
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
		ScoringShadowOperationCreate,
		mode,
		scoringInput,
		req.tracking.ComputedScore,
	)
	if s.useScoringEngine {
		if scoringErr != nil {
			return nil, fmt.Errorf("could not score log: %w", scoringErr)
		}
		ApplyScoringResult(&req.tracking, result)
		for _, registrationID := range req.RegistrationIDs {
			registration := validContestRegistrations[registrationID]
			contestTracking, contestErr := resolveContestLogTracking(
				ctx,
				s.repo,
				registration.ContestID,
				registrationID,
				req.tracking,
				scoringInput,
			)
			if contestErr != nil {
				return nil, fmt.Errorf("could not score contest %s: %w", registration.ContestID, contestErr)
			}
			req.contestTrackings = append(req.contestTrackings, contestTracking)
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

	if err := hydrateLogActivity(log); err != nil {
		return nil, err
	}

	return log, nil
}

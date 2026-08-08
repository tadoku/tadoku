package domain

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type ScorePreviewRepository interface {
	FetchOngoingContestRegistrations(context.Context, *RegistrationListOngoingRequest) (*ContestRegistrations, error)
	FindUnitForTracking(context.Context, *UnitFindForTrackingRequest) (*Unit, error)
	FindUnitForTrackingByKey(context.Context, *UnitFindForTrackingByKeyRequest) (*Unit, error)
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
	FindContestScoringRuleSets(context.Context, uuid.UUID) (*ScoringRuleSet, *ScoringRuleSet, error)
}

type ScorePreviewRequest struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32  `validate:"required"`
	LanguageCode    string `validate:"required"`
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
}

type ScoreEstimate struct {
	Score     float32
	Source    ScoreSource
	RuleSetID *uuid.UUID
	Rules     []AppliedScoringRule
}

type ContestScoreEstimate struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Estimate       ScoreEstimate
}

type ScorePreviewResult struct {
	Platform ScoreEstimate
	Contests []ContestScoreEstimate
}

type ScorePreview struct {
	repo     ScorePreviewRepository
	clock    commondomain.Clock
	validate *validator.Validate
}

func NewScorePreview(repo ScorePreviewRepository, clock commondomain.Clock) *ScorePreview {
	return &ScorePreview{
		repo:     repo,
		clock:    clock,
		validate: validator.New(),
	}
}

func (s *ScorePreview) Execute(ctx context.Context, req *ScorePreviewRequest) (*ScorePreviewResult, error) {
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("unable to validate score preview: %w", ErrInvalidLog)
	}
	var err error
	req.Tags, err = ValidateAndNormalizeTags(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("unable to validate tags: %w", err)
	}

	registrations, err := s.resolveRegistrations(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	tracking, err := resolveLogTracking(
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
	input := ScoringInput{
		ActivityID:      req.ActivityID,
		UnitKey:         tracking.UnitKey,
		LanguageCode:    req.LanguageCode,
		Tags:            req.Tags,
		Amount:          req.Amount,
		DurationSeconds: req.DurationSeconds,
	}
	platformResult, err := EvaluateActivePlatformScore(ctx, s.repo, input)
	if err != nil {
		return nil, fmt.Errorf("could not preview platform score: %w", err)
	}
	ApplyScoringResult(&tracking, platformResult)

	result := &ScorePreviewResult{
		Platform: scoreEstimateFromTracking(tracking),
		Contests: make([]ContestScoreEstimate, 0, len(req.RegistrationIDs)),
	}
	for _, registrationID := range req.RegistrationIDs {
		registration := registrations[registrationID]
		contestTracking, scoringErr := resolveContestLogTracking(
			ctx,
			s.repo,
			registration.ContestID,
			registrationID,
			tracking,
			input,
		)
		if scoringErr != nil {
			return nil, fmt.Errorf("could not preview contest %s score: %w", registration.ContestID, scoringErr)
		}
		result.Contests = append(result.Contests, ContestScoreEstimate{
			RegistrationID: registrationID,
			ContestID:      registration.ContestID,
			Estimate:       scoreEstimateFromTracking(contestTracking.Tracking),
		})
	}
	return result, nil
}

func (s *ScorePreview) resolveRegistrations(
	ctx context.Context,
	userID uuid.UUID,
	req *ScorePreviewRequest,
) (map[uuid.UUID]ContestRegistration, error) {
	result := make(map[uuid.UUID]ContestRegistration, len(req.RegistrationIDs))
	if len(req.RegistrationIDs) == 0 {
		return result, nil
	}
	registrations, err := s.repo.FetchOngoingContestRegistrations(ctx, &RegistrationListOngoingRequest{
		UserID: userID,
		Now:    s.clock.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch registrations: %w", err)
	}
	for _, registration := range registrations.Registrations {
		result[registration.ID] = registration
	}
	for _, registrationID := range req.RegistrationIDs {
		registration, ok := result[registrationID]
		if !ok {
			return nil, fmt.Errorf("registration is not found as ongoing for the current user: %w", ErrInvalidLog)
		}
		if !registrationAllowsLog(registration, req.LanguageCode, req.ActivityID) {
			return nil, fmt.Errorf("log is not allowed for registration: %w", ErrInvalidLog)
		}
	}
	return result, nil
}

func registrationAllowsLog(registration ContestRegistration, languageCode string, activityID int32) bool {
	languageAllowed := false
	for _, language := range registration.Languages {
		if language.Code == languageCode {
			languageAllowed = true
			break
		}
	}
	if !languageAllowed {
		return false
	}
	for _, activity := range registration.Contest.AllowedActivities {
		if activity.ID == activityID {
			return true
		}
	}
	return false
}

func scoreEstimateFromTracking(tracking LogTracking) ScoreEstimate {
	estimate := ScoreEstimate{Score: tracking.ComputedScore}
	if tracking.ScoreProvenance == nil {
		return estimate
	}
	estimate.Source = tracking.ScoreProvenance.Source
	estimate.RuleSetID = tracking.ScoreProvenance.RuleSetID
	estimate.Rules = make([]AppliedScoringRule, len(tracking.ScoreProvenance.RuleIDs))
	for i, ruleID := range tracking.ScoreProvenance.RuleIDs {
		estimate.Rules[i] = AppliedScoringRule{
			RuleID: ruleID,
			Rate:   tracking.ScoreProvenance.Rates[i],
		}
	}
	return estimate
}

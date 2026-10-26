package scoring

import (
	"context"
	"errors"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Service struct {
	repository        *ScoringRepository
	logScoringEnabled bool
	observer          *observability.ScoringObserver
}

func NewService(repository *ScoringRepository, logScoringEnabled bool, observer *observability.ScoringObserver) *Service {
	return &Service{repository: repository, logScoringEnabled: logScoringEnabled, observer: observer}
}

type LogOperation string

const (
	LogCreate LogOperation = "create"
	LogUpdate LogOperation = "update"
)

func (s *Service) ScoreLog(ctx context.Context, operation LogOperation, input logscore.Input, targets []logscore.Target) (logscore.Result, error) {
	input, err := s.NormalizeLogInput(input)
	if err != nil {
		return logscore.Result{}, err
	}

	base, err := s.resolveLogTracking(ctx, input)
	if err != nil {
		return logscore.Result{}, err
	}
	tags := input.Tags
	parameters := PreviewParameters{
		UnitID:          input.UnitID,
		UnitKey:         input.UnitKey,
		ActivityID:      input.ActivityID,
		LanguageCode:    input.LanguageCode,
		Amount:          input.Amount,
		DurationSeconds: input.DurationSeconds,
		Tags:            tags,
	}
	result := logscore.Result{
		ActivityID:   input.ActivityID,
		LanguageCode: input.LanguageCode,
		Tags:         tags,
		Tracking:     base,
	}
	for _, target := range targets {
		result.EligibleOfficial = result.EligibleOfficial || target.Official
	}

	platform, matched, scoringErr := s.ScorePlatform(ctx, parameters)
	mode := "shadow"
	if s.logScoringEnabled {
		mode = "authoritative"
	}
	comparison := observability.ScoringComparison{
		Operation:    string(operation),
		Mode:         mode,
		ActivityID:   input.ActivityID,
		UnitKey:      base.UnitKey,
		LanguageCode: input.LanguageCode,
		LegacyScore:  base.Score,
		Matched:      matched,
	}
	if input.Amount != nil {
		comparison.ScoreSource = "amount"
	} else {
		comparison.ScoreSource = "duration_minutes"
	}
	if scoringErr != nil {
		comparison.ErrorType = scoringErrorType(scoringErr)
	} else {
		comparison.EngineScore = &platform.Score
		comparison.RuleSetID = platform.RuleSetID
		for _, rule := range platform.Rules {
			comparison.AppliedRuleIDs = append(comparison.AppliedRuleIDs, rule.RuleID)
		}
	}
	s.observer.Observe(ctx, comparison)
	if scoringErr != nil && s.logScoringEnabled {
		return logscore.Result{}, scoringErr
	}
	if !s.logScoringEnabled {
		for _, target := range targets {
			result.ContestTrackings = append(result.ContestTrackings, logscore.ContestTracking{
				RegistrationID: target.RegistrationID,
				ContestID:      target.ContestID,
				Tracking:       base,
			})
		}
		return result, nil
	}

	result.Tracking = trackingFromEstimate(base, platform)
	for _, target := range targets {
		estimate, err := s.ScoreContest(ctx, parameters, target.ContestID, platform)
		if err != nil {
			return logscore.Result{}, err
		}
		result.ContestTrackings = append(result.ContestTrackings, logscore.ContestTracking{
			RegistrationID: target.RegistrationID,
			ContestID:      target.ContestID,
			Tracking:       trackingFromEstimate(result.Tracking, estimate),
		})
	}
	return result, nil
}

func (s *Service) NormalizeLogInput(input logscore.Input) (logscore.Input, error) {
	if input.ActivityID == 0 {
		return logscore.Input{}, errx.NewInvalidInputError("activity_id is required")
	}
	if input.LanguageCode == "" {
		return logscore.Input{}, errx.NewInvalidInputError("language_code is required")
	}
	tags, err := normalizeTags(input.Tags)
	if err != nil {
		return logscore.Input{}, err
	}
	input.Tags = tags
	return input, nil
}

func (s *Service) resolveLogTracking(ctx context.Context, input logscore.Input) (logscore.Tracking, error) {
	if !validActivity(input.ActivityID) {
		return logscore.Tracking{}, errx.NewInvalidInputError("invalid log activity")
	}
	legacyDurationRate, _ := activities.LegacyDurationScorePerMinute(input.ActivityID)
	unitActivity, knownUnit := activities.UnitActivityID(stringValue(input.UnitKey))
	if input.UnitKey != nil && (!knownUnit || unitActivity != input.ActivityID) {
		return logscore.Tracking{}, errx.NewInvalidInputError("unit_key is not valid for activity_id")
	}
	var unit *logUnit
	var err error
	if input.Amount != nil {
		switch {
		case input.UnitID != nil:
			unit, err = s.repository.FindLogUnitByID(ctx, *input.UnitID, input.ActivityID, input.LanguageCode)
		case input.UnitKey != nil:
			unit, err = s.repository.FindLogUnitByKey(ctx, *input.UnitKey, input.ActivityID, input.LanguageCode)
		}
		if err != nil {
			return logscore.Tracking{}, err
		}
	}
	if unit != nil {
		unitActivity, knownUnit = activities.UnitActivityID(unit.Key)
		if !knownUnit || unitActivity != input.ActivityID {
			return logscore.Tracking{}, errx.NewInvalidInputError("resolved unit is not valid for activity_id")
		}
	}
	if unit != nil && input.UnitKey != nil && unit.Key != *input.UnitKey {
		return logscore.Tracking{}, errx.NewInvalidInputError("unit_id and unit_key identify different units")
	}
	hasAmount := input.Amount != nil
	hasUnit := input.UnitID != nil || input.UnitKey != nil
	if input.DurationSeconds != nil && *input.DurationSeconds <= 0 {
		return logscore.Tracking{}, errx.NewInvalidInputError("duration_seconds must be positive")
	}
	if input.Amount != nil && (!finite(*input.Amount) || *input.Amount <= 0) {
		return logscore.Tracking{}, errx.NewInvalidInputError("amount must be positive")
	}
	if hasAmount != hasUnit {
		return logscore.Tracking{}, errx.NewInvalidInputError("amount and a unit identifier must be supplied together")
	}
	if !hasAmount && input.DurationSeconds == nil {
		return logscore.Tracking{}, errx.NewInvalidInputError("amount/unit or duration_seconds is required")
	}
	tracking := logscore.Tracking{DurationSeconds: input.DurationSeconds}
	if hasAmount {
		if unit == nil {
			return logscore.Tracking{}, errx.NewInvalidInputError("unit is required for amount scoring")
		}
		tracking.UnitID = &unit.ID
		tracking.UnitKey = unit.Key
		tracking.Amount = input.Amount
		tracking.Modifier = &unit.Modifier
		tracking.Score = *input.Amount * unit.Modifier
	} else {
		tracking.Score = float32(*input.DurationSeconds) / 60 * legacyDurationRate
	}
	return tracking, nil
}

func scoringErrorType(err error) string {
	if errors.Is(err, ErrRuleSetNotFound) {
		return "scoring_rule_set_not_found"
	}
	switch errx.KindOf(err) {
	case errx.InvalidInput:
		return "invalid_scoring_input"
	case errx.Internal:
		return "invalid_scoring_rule_set"
	default:
		return "evaluation_failed"
	}
}

func trackingFromEstimate(base logscore.Tracking, estimate Estimate) logscore.Tracking {
	base.Score = estimate.Score
	base.RuleSetID = estimate.RuleSetID
	base.Source = string(estimate.Source)
	base.RuleIDs = nil
	base.Rates = nil
	for _, rule := range estimate.Rules {
		base.RuleIDs = append(base.RuleIDs, rule.RuleID)
		base.Rates = append(base.Rates, rule.Rate)
	}
	return base
}

func (s *Service) ScoreLogAttachments(ctx context.Context, log logscore.Input, base logscore.Tracking, targets []logscore.Target) ([]logscore.ContestTracking, error) {
	attachments := make([]logscore.ContestTracking, 0, len(targets))
	for _, target := range targets {
		tracking := base
		if s.logScoringEnabled {
			unitKey := base.UnitKey
			estimate, err := s.ScoreResolvedContest(ctx, PreviewParameters{
				UnitKey:         &unitKey,
				ActivityID:      log.ActivityID,
				LanguageCode:    log.LanguageCode,
				Amount:          base.Amount,
				DurationSeconds: base.DurationSeconds,
				Tags:            log.Tags,
			}, target.ContestID)
			if err != nil {
				return nil, err
			}
			if estimate != nil {
				tracking = trackingFromEstimate(base, *estimate)
			}
		}
		attachments = append(attachments, logscore.ContestTracking{
			RegistrationID: target.RegistrationID,
			ContestID:      target.ContestID,
			Tracking:       tracking,
		})
	}
	return attachments, nil
}

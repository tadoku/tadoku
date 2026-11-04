package scoring

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *Service) Preview(ctx context.Context, parameters PreviewParameters) (*Preview, error) {
	unitKey, err := resolveUnitKey(parameters)
	if err != nil {
		return nil, err
	}
	tags, err := normalizeTags(parameters.Tags)
	if err != nil {
		return nil, err
	}

	input := scoringInput{
		activityID:      parameters.ActivityID,
		unitKey:         unitKey,
		languageCode:    parameters.LanguageCode,
		tags:            tags,
		amount:          parameters.Amount,
		durationSeconds: parameters.DurationSeconds,
	}

	platformSet, err := s.repository.FindActivePlatformRuleSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("preview platform score: %w", err)
	}
	platformSet.Rules, err = s.repository.ListRules(ctx, platformSet.ID)
	if err != nil {
		return nil, err
	}
	platform, _, err := evaluate(input, *platformSet)
	if err != nil {
		return nil, fmt.Errorf("preview platform score: %w", err)
	}

	result := &Preview{
		Platform: platform,
		Contests: make([]ContestEstimate, 0, len(parameters.Contests)),
	}
	for _, contest := range parameters.Contests {
		set, fallback, err := s.findContestRuleSets(ctx, contest.ContestID)
		if err != nil {
			return nil, fmt.Errorf("preview contest score: %w", err)
		}
		estimate := platform
		if set != nil {
			var matched bool
			estimate, matched, err = evaluate(input, *set)
			if err != nil {
				return nil, fmt.Errorf("preview contest score: %w", err)
			}
			if !matched && set.Mode == "override" {
				if fallback == nil {
					return nil, errx.NewInternalError("override scoring rule set requires a fallback")
				}
				estimate, _, err = evaluate(input, *fallback)
			}
			if !matched && set.Mode != "override" && set.Mode != "replace" {
				return nil, errx.NewInternalError("unknown contest scoring mode")
			}
			if err != nil {
				return nil, fmt.Errorf("preview contest score: %w", err)
			}
		}
		result.Contests = append(result.Contests, ContestEstimate{
			RegistrationID: contest.RegistrationID,
			ContestID:      contest.ContestID,
			Estimate:       estimate,
		})
	}
	return result, nil
}

func (s *Service) ScorePlatform(ctx context.Context, parameters PreviewParameters) (Estimate, bool, error) {
	unitKey, err := resolveUnitKey(parameters)
	if err != nil {
		return Estimate{}, false, err
	}
	input := scoringInput{
		activityID:      parameters.ActivityID,
		unitKey:         unitKey,
		languageCode:    parameters.LanguageCode,
		tags:            parameters.Tags,
		amount:          parameters.Amount,
		durationSeconds: parameters.DurationSeconds,
	}
	set, err := s.repository.FindActivePlatformRuleSet(ctx)
	if err != nil {
		return Estimate{}, false, err
	}
	set.Rules, err = s.repository.ListRules(ctx, set.ID)
	if err != nil {
		return Estimate{}, false, err
	}
	return evaluate(input, *set)
}

func (s *Service) ScoreContest(ctx context.Context, parameters PreviewParameters, contestID uuid.UUID, platform Estimate) (Estimate, error) {
	unitKey, err := resolveUnitKey(parameters)
	if err != nil {
		return Estimate{}, err
	}
	input := scoringInput{
		activityID:      parameters.ActivityID,
		unitKey:         unitKey,
		languageCode:    parameters.LanguageCode,
		tags:            parameters.Tags,
		amount:          parameters.Amount,
		durationSeconds: parameters.DurationSeconds,
	}
	estimate, err := s.scoreContest(ctx, input, contestID)
	if estimate == nil || err != nil {
		return platform, err
	}
	return *estimate, nil
}

func (s *Service) ScoreResolvedContest(ctx context.Context, parameters PreviewParameters, contestID uuid.UUID) (*Estimate, error) {
	input := scoringInput{
		activityID:      parameters.ActivityID,
		unitKey:         stringValue(parameters.UnitKey),
		languageCode:    parameters.LanguageCode,
		tags:            parameters.Tags,
		amount:          parameters.Amount,
		durationSeconds: parameters.DurationSeconds,
	}

	return s.scoreContest(ctx, input, contestID)
}

func (s *Service) scoreContest(ctx context.Context, input scoringInput, contestID uuid.UUID) (*Estimate, error) {
	set, fallback, err := s.findContestRuleSets(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, nil
	}
	estimate, matched, err := evaluate(input, *set)
	if err != nil {
		return nil, err
	}
	if !matched && set.Mode == "override" {
		if fallback == nil {
			return nil, errx.NewInternalError("override scoring rule set requires a fallback")
		}
		estimate, _, err = evaluate(input, *fallback)
	}
	if !matched && set.Mode != "override" && set.Mode != "replace" {
		return nil, errx.NewInternalError("unknown contest scoring mode")
	}
	return &estimate, err
}

func resolveUnit(parameters PreviewParameters) (*activities.Unit, error) {
	if !validActivity(parameters.ActivityID) {
		return nil, errx.NewInvalidInputError("activity_id is not valid")
	}
	if (parameters.Amount != nil) != (parameters.UnitKey != nil) {
		return nil, errx.NewInvalidInputError("amount and unit_key must be supplied together")
	}
	if parameters.DurationSeconds != nil && *parameters.DurationSeconds <= 0 {
		return nil, errx.NewInvalidInputError("duration_seconds must be positive")
	}
	if parameters.Amount != nil && (!finite(*parameters.Amount) || *parameters.Amount <= 0) {
		return nil, errx.NewInvalidInputError("amount must be positive and finite")
	}
	if parameters.Amount == nil && parameters.DurationSeconds == nil {
		return nil, errx.NewInvalidInputError("amount/unit_key or duration_seconds is required")
	}
	if parameters.Amount == nil {
		return nil, nil
	}

	unit, err := activities.ResolveUnit(parameters.ActivityID, *parameters.UnitKey, parameters.LanguageCode)
	if err != nil {
		return nil, errx.NewInvalidInputError(err.Error())
	}
	return &unit, nil
}

func resolveUnitKey(parameters PreviewParameters) (string, error) {
	unit, err := resolveUnit(parameters)
	if err != nil || unit == nil {
		return "", err
	}
	return unit.Key, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) findContestRuleSets(ctx context.Context, contestID uuid.UUID) (*RuleSet, *RuleSet, error) {
	id, err := s.repository.FindContestActiveRuleSetID(ctx, contestID)
	if err != nil || id == nil {
		return nil, nil, err
	}
	set, err := s.repository.FindRuleSetByID(ctx, *id)
	if err != nil {
		return nil, nil, err
	}
	set.Rules, err = s.repository.ListRules(ctx, set.ID)
	if err != nil || set.FallbackRuleSetID == nil {
		return set, nil, err
	}
	fallback, err := s.repository.FindRuleSetByID(ctx, *set.FallbackRuleSetID)
	if err != nil {
		return nil, nil, err
	}
	fallback.Rules, err = s.repository.ListRules(ctx, fallback.ID)
	return set, fallback, err
}

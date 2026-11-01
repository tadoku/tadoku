package scoring

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *Service) Preview(ctx context.Context, parameters PreviewParameters) (*Preview, error) {
	unitKey, err := s.resolveUnit(ctx, parameters)
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
	unitKey, err := s.resolveUnit(ctx, parameters)
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
	unitKey, err := s.resolveUnit(ctx, parameters)
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

func (s *Service) resolveUnit(ctx context.Context, parameters PreviewParameters) (string, error) {
	if !validActivity(parameters.ActivityID) {
		return "", errx.NewInvalidInputError("activity_id is not valid")
	}
	unitActivity, knownUnit := activities.UnitActivityID(stringValue(parameters.UnitKey))
	if parameters.UnitKey != nil && (!knownUnit || unitActivity != parameters.ActivityID) {
		return "", errx.NewInvalidInputError("unit_key is not valid for activity_id")
	}
	hasUnit := parameters.UnitID != nil || parameters.UnitKey != nil
	if (parameters.Amount != nil) != hasUnit {
		return "", errx.NewInvalidInputError("amount and a unit identifier must be supplied together")
	}
	if parameters.DurationSeconds != nil && *parameters.DurationSeconds <= 0 {
		return "", errx.NewInvalidInputError("duration_seconds must be positive")
	}
	if parameters.Amount != nil && (!finite(*parameters.Amount) || *parameters.Amount <= 0) {
		return "", errx.NewInvalidInputError("amount must be positive and finite")
	}
	if parameters.Amount == nil && parameters.DurationSeconds == nil {
		return "", errx.NewInvalidInputError("amount/unit or duration_seconds is required")
	}
	if parameters.Amount == nil {
		return "", nil
	}

	unit, err := s.findApplicableUnit(
		ctx,
		parameters.UnitID,
		parameters.UnitKey,
		parameters.ActivityID,
		parameters.LanguageCode,
	)
	if err != nil {
		return "", err
	}
	if unit == nil && parameters.UnitID != nil {
		return "", errx.NewInvalidInputError("unit_id is not valid for activity_id and language_code")
	}
	if unit == nil {
		return "", errx.NewInvalidInputError("unit_key is not valid for activity_id and language_code")
	}
	if parameters.UnitKey != nil && unit.Key != *parameters.UnitKey {
		return "", errx.NewInvalidInputError("unit_id and unit_key identify different units")
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

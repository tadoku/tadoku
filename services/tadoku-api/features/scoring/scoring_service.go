package scoring

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
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
		parameters.Contests = append(parameters.Contests, PreviewContest{
			RegistrationID: target.RegistrationID,
			ContestID:      target.ContestID,
		})
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
	tags, err := NormalizeTags(input.Tags)
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
		unit, err = s.repository.FindLogUnit(ctx, input.UnitID, input.UnitKey, input.ActivityID, input.LanguageCode)
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

func (s *Service) ListPlatformRuleSets(ctx context.Context, includeDrafts bool) ([]RuleSet, error) {
	sets, err := s.repository.ListPlatformRuleSets(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateRuleSets(ctx, sets); err != nil {
		return nil, err
	}
	active, err := s.repository.FindActivePlatformRuleSet(ctx)
	if err != nil {
		return nil, err
	}
	for i := range sets {
		sets[i].Active = sets[i].ID == active.ID
	}
	if includeDrafts {
		return sets, nil
	}
	published := make([]RuleSet, 0, len(sets))
	for _, set := range sets {
		if set.Status == "published" {
			published = append(published, set)
		}
	}
	return published, nil
}

func (s *Service) ListContestRuleSets(ctx context.Context, contestID uuid.UUID) ([]RuleSet, error) {
	sets, err := s.repository.ListContestRuleSets(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateRuleSets(ctx, sets); err != nil {
		return nil, err
	}
	activeID, err := s.repository.FindContestActiveRuleSetID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if activeID != nil {
		for i := range sets {
			sets[i].Active = sets[i].ID == *activeID
		}
	}
	return sets, nil
}

func (s *Service) Preview(ctx context.Context, parameters PreviewParameters) (*Preview, error) {
	unitKey, err := s.resolveUnit(ctx, parameters)
	if err != nil {
		return nil, err
	}
	tags, err := NormalizeTags(parameters.Tags)
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

	var resolved string
	if parameters.UnitID != nil {
		var err error
		resolved, err = s.repository.FindUnitKeyByID(ctx, *parameters.UnitID, parameters.ActivityID, parameters.LanguageCode)
		if err != nil {
			return "", err
		}
	}
	if parameters.UnitID == nil && parameters.UnitKey != nil {
		var err error
		resolved, err = s.repository.FindUnitKeyByKey(ctx, *parameters.UnitKey, parameters.ActivityID, parameters.LanguageCode)
		if err != nil {
			return "", err
		}
	}
	if parameters.UnitKey != nil && resolved != *parameters.UnitKey {
		return "", errx.NewInvalidInputError("unit_id and unit_key identify different units")
	}
	unitActivity, knownUnit = activities.UnitActivityID(resolved)
	if !knownUnit || unitActivity != parameters.ActivityID {
		return "", errx.NewInvalidInputError("resolved unit_key is not valid for activity_id")
	}
	return resolved, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) hydrateRuleSets(ctx context.Context, sets []RuleSet) error {
	for i := range sets {
		rules, err := s.repository.ListRules(ctx, sets[i].ID)
		if err != nil {
			return err
		}
		sets[i].Rules = rules
	}
	return nil
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

func (s *Service) ValidateContestDraftConfiguration(ctx context.Context, parameters *ContestDraftParameters) error {
	var fallbackID uuid.UUID
	switch configuration := parameters.Configuration.(type) {
	case Replace:
		return nil
	case Override:
		fallbackID = configuration.FallbackRuleSetID
	default:
		return errx.NewInternalError("invalid contest draft configuration")
	}

	fallback, err := s.repository.FindRuleSetByID(ctx, fallbackID)
	if err != nil {
		return err
	}
	if fallback.Scope != "platform" || fallback.Status != "published" {
		return errx.NewInvalidInputError("fallback must be a published platform rule set")
	}

	return nil
}

func (s *Service) NormalizeDraftRules(parameters *DraftRules, languages []domainlanguages.Language) error {
	return parameters.normalize(languages)
}

func (s *Service) CreatePlatformDraft(ctx context.Context, parameters PlatformDraftParameters) (*RuleSet, error) {
	return s.createDraft(ctx, RuleSet{
		Scope:     "platform",
		Rules:     parameters.Rules,
		CreatedAt: parameters.CreatedAt,
	})
}

func (s *Service) CreateContestDraft(ctx context.Context, parameters ContestDraftParameters) (*RuleSet, error) {
	draft := RuleSet{
		Scope:     "contest",
		ContestID: &parameters.ContestID,
		Rules:     parameters.Rules,
		CreatedAt: parameters.CreatedAt,
	}
	switch configuration := parameters.Configuration.(type) {
	case Replace:
		draft.Mode = string(ModeReplace)
	case Override:
		draft.Mode = string(ModeOverride)
		draft.FallbackRuleSetID = &configuration.FallbackRuleSetID
	default:
		return nil, errx.NewInternalError("invalid contest draft configuration")
	}

	return s.createDraft(ctx, draft)
}

func (s *Service) createDraft(ctx context.Context, draft RuleSet) (*RuleSet, error) {
	version, err := s.repository.NextDraftVersion(ctx, draft.ContestID)
	if err != nil {
		return nil, fmt.Errorf("allocate scoring rule set version: %w", err)
	}

	draft.ID = uuid.New()
	draft.Version = version
	draft.Status = "draft"
	rules := draft.Rules
	draft.Rules = []Rule{}

	created, err := s.repository.CreateDraft(ctx, draft)
	if err != nil {
		return nil, err
	}
	created.Rules = make([]Rule, len(rules))
	for i, rule := range rules {
		rule.ID = uuid.New()
		if err := s.repository.CreateRule(ctx, created.ID, rule); err != nil {
			return nil, err
		}
		created.Rules[i] = rule
	}

	return created, nil
}

func (s *Service) FindRuleSet(ctx context.Context, id uuid.UUID) (*RuleSet, error) {
	return s.repository.FindRuleSetByID(ctx, id)
}

func (s *Service) PublishRuleSet(ctx context.Context, ruleSet RuleSet, publishedAt time.Time) (*RuleSet, error) {
	if ruleSet.Status != "draft" {
		return nil, errx.NewConflictError("only draft scoring rule sets can be published")
	}

	published, err := s.repository.PublishRuleSet(ctx, ruleSet.ID, publishedAt)
	if err != nil {
		return nil, err
	}
	published.Rules, err = s.repository.ListRules(ctx, published.ID)
	return published, err
}

func (s *Service) ActivateRuleSet(ctx context.Context, ruleSet RuleSet, updatedAt time.Time) error {
	if ruleSet.Status != "published" {
		return errx.NewConflictError("only published scoring rule sets can be activated")
	}

	switch ruleSet.Scope {
	case "platform":
		return s.repository.ActivatePlatformRuleSet(ctx, ruleSet.ID)
	case "contest":
		if ruleSet.ContestID == nil {
			return errx.NewInvalidInputError("contest scoring rule set requires contest_id")
		}
		return s.repository.ActivateContestRuleSet(ctx, *ruleSet.ContestID, ruleSet.ID, updatedAt)
	default:
		return errx.NewInvalidInputError("scoring rule set scope is invalid")
	}
}

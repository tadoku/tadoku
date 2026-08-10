package domain

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
)

type activePlatformScoringRuleSetFinder interface {
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
}

type ScoringShadowOutcome string

const (
	ScoringShadowOutcomeMatch     ScoringShadowOutcome = "match"
	ScoringShadowOutcomeMismatch  ScoringShadowOutcome = "mismatch"
	ScoringShadowOutcomeUnmatched ScoringShadowOutcome = "unmatched"
	ScoringShadowOutcomeError     ScoringShadowOutcome = "error"
)

type ScoringShadowOperation string

const (
	ScoringShadowOperationCreate ScoringShadowOperation = "create"
	ScoringShadowOperationUpdate ScoringShadowOperation = "update"
)

type ScoringShadowMode string

const (
	ScoringShadowModeShadow        ScoringShadowMode = "shadow"
	ScoringShadowModeAuthoritative ScoringShadowMode = "authoritative"
)

// ScoringShadowObservation is the privacy-reviewed data contract passed to the
// telemetry adapter. It intentionally contains no user, registration, or log
// identifiers and no user-authored content.
type ScoringShadowObservation struct {
	Outcome        ScoringShadowOutcome
	Operation      ScoringShadowOperation
	Mode           ScoringShadowMode
	ActivityID     int32
	UnitKey        string
	LanguageCode   string
	ScoreSource    ScoreSource
	LegacyScore    float32
	EngineScore    *float32
	AbsoluteDelta  *float64
	RelativeDelta  *float64
	RuleSetID      *uuid.UUID
	AppliedRuleIDs []uuid.UUID
	ErrorType      string
}

// ScoringShadowObserver is defined where scoring comparisons are consumed so
// the domain does not depend on an observability implementation.
type ScoringShadowObserver interface {
	ObserveScoringShadow(context.Context, ScoringShadowObservation)
}

type ScoringShadowComparison struct {
	InterimScore float32
	RuleResult   ScoringResult
	Mismatch     bool
}

func EvaluatePlatformScoringShadow(
	ctx context.Context,
	finder activePlatformScoringRuleSetFinder,
	input ScoringInput,
	interimScore float32,
) (ScoringShadowComparison, error) {
	ruleResult, err := EvaluateActivePlatformScore(ctx, finder, input)
	if err != nil {
		return ScoringShadowComparison{}, err
	}
	return ScoringShadowComparison{
		InterimScore: interimScore,
		RuleResult:   ruleResult,
		Mismatch:     !scoringScoresEqual(interimScore, ruleResult.Score),
	}, nil
}

func EvaluateActivePlatformScore(
	ctx context.Context,
	finder activePlatformScoringRuleSetFinder,
	input ScoringInput,
) (ScoringResult, error) {
	ruleSet, err := finder.FindActivePlatformScoringRuleSet(ctx)
	if err != nil {
		return ScoringResult{}, err
	}
	if ruleSet == nil {
		return ScoringResult{}, fmt.Errorf("active platform scoring rule set is nil: %w", ErrScoringRuleSetNotFound)
	}
	return EvaluateScoringRuleSet(input, *ruleSet)
}

// EvaluateAndObservePlatformScoring evaluates the engine exactly once and
// records exactly one mutually exclusive outcome. Observer failures cannot
// affect scoring because the observer has no error return.
func EvaluateAndObservePlatformScoring(
	ctx context.Context,
	finder activePlatformScoringRuleSetFinder,
	observer ScoringShadowObserver,
	operation ScoringShadowOperation,
	mode ScoringShadowMode,
	input ScoringInput,
	legacyScore float32,
) (ScoringResult, error) {
	result, err := EvaluateActivePlatformScore(ctx, finder, input)
	observation := ScoringShadowObservation{
		Operation:    operation,
		Mode:         mode,
		ActivityID:   input.ActivityID,
		UnitKey:      input.UnitKey,
		LanguageCode: input.LanguageCode,
		ScoreSource:  scoringShadowScoreSource(input),
		LegacyScore:  legacyScore,
	}
	if err != nil {
		observation.Outcome = ScoringShadowOutcomeError
		observation.ErrorType = scoringShadowErrorType(err)
		observeScoringShadow(ctx, observer, observation)
		return ScoringResult{}, err
	}

	engineScore := result.Score
	absoluteDelta := math.Abs(float64(legacyScore - result.Score))
	scale := math.Max(math.Abs(float64(legacyScore)), math.Abs(float64(result.Score)))
	relativeDelta := float64(0)
	if scale > 0 {
		relativeDelta = absoluteDelta / scale
	}
	observation.EngineScore = &engineScore
	observation.AbsoluteDelta = &absoluteDelta
	observation.RelativeDelta = &relativeDelta
	observation.RuleSetID = result.AppliedRuleSetID
	observation.AppliedRuleIDs = make([]uuid.UUID, len(result.AppliedRules))
	for index, rule := range result.AppliedRules {
		observation.AppliedRuleIDs[index] = rule.RuleID
	}

	switch {
	case !result.Matched:
		observation.Outcome = ScoringShadowOutcomeUnmatched
	case scoringScoresEqual(legacyScore, result.Score):
		observation.Outcome = ScoringShadowOutcomeMatch
	default:
		observation.Outcome = ScoringShadowOutcomeMismatch
	}
	observeScoringShadow(ctx, observer, observation)
	return result, nil
}

func observeScoringShadow(ctx context.Context, observer ScoringShadowObserver, observation ScoringShadowObservation) {
	if observer != nil {
		observer.ObserveScoringShadow(ctx, observation)
	}
}

func scoringShadowScoreSource(input ScoringInput) ScoreSource {
	if input.Amount != nil {
		return ScoreSourceAmount
	}
	if input.DurationSeconds != nil {
		return ScoreSourceDurationMinutes
	}
	return ""
}

func scoringShadowErrorType(err error) string {
	switch {
	case errors.Is(err, ErrScoringRuleSetNotFound):
		return "scoring_rule_set_not_found"
	case errors.Is(err, ErrInvalidScoringRuleSet):
		return "invalid_scoring_rule_set"
	case errors.Is(err, ErrInvalidLog):
		return "invalid_scoring_input"
	default:
		return "evaluation_failed"
	}
}

func scoringScoresEqual(left, right float32) bool {
	difference := math.Abs(float64(left - right))
	scale := math.Max(1, math.Max(math.Abs(float64(left)), math.Abs(float64(right))))
	return difference <= 0.00001*scale
}

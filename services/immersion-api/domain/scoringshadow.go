package domain

import (
	"context"
	"fmt"
	"log/slog"
	"math"
)

type activePlatformScoringRuleSetFinder interface {
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
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
	ruleSet, err := finder.FindActivePlatformScoringRuleSet(ctx)
	if err != nil {
		return ScoringShadowComparison{}, err
	}
	if ruleSet == nil {
		return ScoringShadowComparison{}, fmt.Errorf("active platform scoring rule set is nil: %w", ErrScoringRuleSetNotFound)
	}
	ruleResult, err := EvaluateScoringRuleSet(input, *ruleSet)
	if err != nil {
		return ScoringShadowComparison{}, err
	}
	return ScoringShadowComparison{
		InterimScore: interimScore,
		RuleResult:   ruleResult,
		Mismatch:     !scoringScoresEqual(interimScore, ruleResult.Score),
	}, nil
}

func RecordPlatformScoringShadow(
	ctx context.Context,
	finder activePlatformScoringRuleSetFinder,
	input ScoringInput,
	interimScore float32,
) {
	comparison, err := EvaluatePlatformScoringShadow(ctx, finder, input, interimScore)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"scoring shadow evaluation failed",
			"activity_id", input.ActivityID,
			"unit_key", input.UnitKey,
			"language_code", input.LanguageCode,
			"error", err,
		)
		return
	}
	if !comparison.RuleResult.Matched {
		slog.WarnContext(
			ctx,
			"scoring shadow input unmatched",
			"activity_id", input.ActivityID,
			"unit_key", input.UnitKey,
			"language_code", input.LanguageCode,
			"score_source", comparison.RuleResult.ScoreSource,
			"interim_score", interimScore,
		)
		return
	}
	if comparison.Mismatch {
		slog.WarnContext(
			ctx,
			"scoring shadow mismatch",
			"activity_id", input.ActivityID,
			"unit_key", input.UnitKey,
			"language_code", input.LanguageCode,
			"score_source", comparison.RuleResult.ScoreSource,
			"interim_score", interimScore,
			"rule_score", comparison.RuleResult.Score,
			"rule_set_id", comparison.RuleResult.AppliedRuleSetID,
		)
	}
}

func scoringScoresEqual(left, right float32) bool {
	difference := math.Abs(float64(left - right))
	scale := math.Max(1, math.Max(math.Abs(float64(left)), math.Abs(float64(right))))
	return difference <= 0.00001*scale
}

package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type contestScoringRuleSetFinder interface {
	FindActivePlatformScoringRuleSet(context.Context) (*ScoringRuleSet, error)
	FindContestScoringRuleSets(context.Context, uuid.UUID) (*ScoringRuleSet, *ScoringRuleSet, error)
}

type ContestLogTracking struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Tracking       LogTracking
}

func scoringInputFromTracking(
	activityID int32,
	languageCode string,
	tags []string,
	tracking LogTracking,
) ScoringInput {
	input := ScoringInput{
		ActivityID:   activityID,
		UnitKey:      tracking.UnitKey,
		LanguageCode: languageCode,
		Tags:         tags,
	}
	if tracking.Kind == LogTrackingAmountUnit || tracking.Kind == LogTrackingBoth {
		input.Amount = &tracking.Amount
	}
	if tracking.Kind == LogTrackingDuration || tracking.Kind == LogTrackingBoth {
		input.DurationSeconds = &tracking.DurationSeconds
	}
	return input
}

func resolveContestLogTracking(
	ctx context.Context,
	finder contestScoringRuleSetFinder,
	contestID uuid.UUID,
	registrationID uuid.UUID,
	platformTracking LogTracking,
	input ScoringInput,
) (ContestLogTracking, error) {
	contestRuleSet, fallbackRuleSet, err := finder.FindContestScoringRuleSets(ctx, contestID)
	if err != nil {
		return ContestLogTracking{}, err
	}

	tracking := platformTracking
	if contestRuleSet == nil {
		result, evaluateErr := EvaluateActivePlatformScore(ctx, finder, input)
		if evaluateErr != nil {
			return ContestLogTracking{}, fmt.Errorf("could not evaluate platform scoring rules: %w", evaluateErr)
		}
		ApplyScoringResult(&tracking, result)
	} else {
		result, evaluateErr := EvaluateContestScore(input, *contestRuleSet, fallbackRuleSet)
		if evaluateErr != nil {
			return ContestLogTracking{}, fmt.Errorf("could not evaluate contest scoring rules: %w", evaluateErr)
		}
		ApplyScoringResult(&tracking, result)
	}

	return ContestLogTracking{
		RegistrationID: registrationID,
		ContestID:      contestID,
		Tracking:       tracking,
	}, nil
}

package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
)

type ContestProfileScores struct {
	Registration *ContestRegistration
	Scores       []logs.Score
	OverallScore float32
}

func (a *Application) ContestProfileScores(ctx context.Context, userID, contestID uuid.UUID) (*ContestProfileScores, error) {
	registration, err := a.contests.FindRegistrationWithContest(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	// Legacy registration hydration retains a slot for every stored language code,
	// even when codes repeat or no longer have a language row.
	for len(registration.Languages) < len(registration.LanguageCodes) {
		registration.Languages = append(registration.Languages, contests.Language{})
	}
	scores, err := a.logs.ContestScores(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	result := &ContestProfileScores{
		Registration: registration,
		Scores:       scores,
	}
	for _, score := range scores {
		result.OverallScore += score.Score
	}
	return result, nil
}

func (a *Application) ContestProfileActivity(ctx context.Context, userID, contestID uuid.UUID) ([]logs.ContestActivity, error) {
	return a.logs.ContestActivity(ctx, userID, contestID)
}

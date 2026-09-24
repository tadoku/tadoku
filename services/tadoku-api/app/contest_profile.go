package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
)

type ContestProfileScores struct {
	Registration *ContestRegistration
	Scores       []logs.Score
	OverallScore float32
}

func (a *Application) ContestProfileScores(ctx context.Context, userID, contestID uuid.UUID) (*ContestProfileScores, error) {
	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	registration, err := a.contests.FindRegistrationWithContest(ctx, userID, contestID, languages)
	if err != nil {
		return nil, err
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

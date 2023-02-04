package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

func (r *Repository) FetchContestSummary(ctx context.Context, req *query.FetchContestSummaryRequest) (*query.FetchContestSummaryResponse, error) {
	summary, err := r.q.ContestSummary(ctx, req.ContestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch contest summary: %w", err)
	}

	return &query.FetchContestSummaryResponse{
		ParticipantCount: int(summary.ParticipantCount),
		LanguageCount:    int(summary.LanguageCount),
		TotalScore:       summary.TotalScore,
	}, nil
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

func (r *Repository) FetchOfficialLeaderboardPreviewForYear(ctx context.Context, req *query.FetchOfficialLeaderboardPreviewForYearRequest) (*query.Leaderboard, error) {
	entries, err := r.q.OfficialLeaderboardPreviewForYear(ctx, int16(req.Year))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.Leaderboard{
				Entries:       []query.LeaderboardEntry{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch leaderboard preview for year: %w", err)
	}

	res := make([]query.LeaderboardEntry, len(entries))
	for i, e := range entries {
		res[i] = query.LeaderboardEntry{
			Rank:            int(e.Rank),
			UserID:          e.UserID,
			UserDisplayName: e.UserDisplayName,
			Score:           e.Score,
			IsTie:           e.IsTie,
		}
	}

	return &query.Leaderboard{
		Entries: res,
	}, nil
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchGlobalLeaderboard(ctx context.Context, req *query.FetchGlobalLeaderboardRequest) (*query.Leaderboard, error) {
	entries, err := r.q.GlobalLeaderboard(ctx, postgres.GlobalLeaderboardParams{
		LanguageCode: postgres.NewNullString(req.LanguageCode),
		ActivityID:   postgres.NewNullInt32(req.ActivityID),
		StartFrom:    int32(req.Page * req.PageSize),
		PageSize:     int32(req.PageSize),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.Leaderboard{
				Entries:       []query.LeaderboardEntry{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch leaderboard: %w", err)
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

	var totalSize int64
	if len(entries) > 0 {
		totalSize = entries[0].TotalSize
	}
	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < int(totalSize) {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &query.Leaderboard{
		Entries:       res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

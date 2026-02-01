package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchContestLeaderboard(ctx context.Context, req *domain.ContestLeaderboardFetchRequest) (*domain.Leaderboard, error) {
	_, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	entries, err := r.q.LeaderboardForContest(ctx, postgres.LeaderboardForContestParams{
		ContestID:    req.ContestID,
		LanguageCode: postgres.NewNullString(req.LanguageCode),
		ActivityID:   postgres.NewNullInt32(req.ActivityID),
		StartFrom:    int32(req.Page * req.PageSize),
		PageSize:     int32(req.PageSize),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.Leaderboard{
				Entries:       []domain.LeaderboardEntry{},
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch leaderboard for contest: %w", err)
	}

	res := make([]domain.LeaderboardEntry, len(entries))
	for i, e := range entries {
		res[i] = domain.LeaderboardEntry{
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

	return &domain.Leaderboard{
		Entries:       res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

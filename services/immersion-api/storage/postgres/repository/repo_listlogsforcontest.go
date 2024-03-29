package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ListLogsForContest(ctx context.Context, req *query.ListLogsForContestRequest) (*query.ListLogsForContestResponse, error) {
	_, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	entries, err := r.q.ListLogsForContest(ctx, postgres.ListLogsForContestParams{
		ContestID:      req.ContestID,
		UserID:         req.UserID,
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.ListLogsForContestResponse{
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	res := make([]query.Log, len(entries))
	for i, it := range entries {
		it := it
		res[i] = query.Log{
			ID:              it.ID,
			UserID:          it.UserID,
			UserDisplayName: &it.UserDisplayName,
			Description:     postgres.NewStringFromNullString(it.Description),
			LanguageCode:    it.LanguageCode,
			LanguageName:    it.LanguageName,
			ActivityID:      int(it.ActivityID),
			ActivityName:    it.ActivityName,
			UnitName:        it.UnitName,
			Tags:            it.Tags,
			Amount:          it.Amount,
			Modifier:        it.Modifier,
			Score:           it.Score,
			CreatedAt:       it.CreatedAt,
			UpdatedAt:       it.UpdatedAt,
			Deleted:         it.DeletedAt.Valid,
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

	return &query.ListLogsForContestResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

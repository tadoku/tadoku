package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ListLogsForContest(ctx context.Context, req *domain.LogListForContestRequest) (*domain.LogListForContestResponse, error) {
	_, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
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
			return &domain.LogListForContestResponse{
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	res := make([]domain.Log, len(entries))
	for i, it := range entries {
		// UnitName from ListLogsForContestRow is string (not nullable per sqlc)
		var unitName *string
		if it.UnitName != "" {
			unitName = &it.UnitName
		}

		res[i] = domain.Log{
			ID:              it.ID,
			UserID:          it.UserID,
			UserDisplayName: &it.UserDisplayName,
			Description:     postgres.NewStringFromNullString(it.Description),
			LanguageCode:    it.LanguageCode,
			LanguageName:    it.LanguageName,
			ActivityID:      int(it.ActivityID),
			ActivityName:    it.ActivityName,
			UnitName:        unitName,
			Tags:            postgres.StringArrayFromInterface(it.Tags),
			Amount:          postgres.NewFloat32FromNullFloat64(it.Amount),
			Modifier:        postgres.NewFloat32FromNullFloat64(it.Modifier),
			Score:           float32(it.Score.Float64),
			DurationSeconds: postgres.NewInt32FromNullInt32(it.DurationSeconds),
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

	return &domain.LogListForContestResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

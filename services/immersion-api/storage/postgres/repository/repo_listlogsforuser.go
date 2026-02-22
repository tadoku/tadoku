package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ListLogsForUser(ctx context.Context, req *domain.LogListForUserRequest) (*domain.LogListForUserResponse, error) {
	entries, err := r.q.ListLogsForUser(ctx, postgres.ListLogsForUserParams{
		UserID:         req.UserID,
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.LogListForUserResponse{
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	res := make([]domain.Log, len(entries))
	for i, it := range entries {
		res[i] = domain.Log{
			ID:              it.ID,
			UserID:          it.UserID,
			Description:     postgres.NewStringFromNullString(it.Description),
			LanguageCode:    it.LanguageCode,
			LanguageName:    it.LanguageName,
			ActivityID:        int(it.ActivityID),
			ActivityName:      it.ActivityName,
			ActivityInputType: it.ActivityInputType,
			UnitName:        postgres.NewStringFromNullString(it.UnitName),
			Tags:            postgres.StringArrayFromInterface(it.Tags),
			Amount:          postgres.NewFloat32FromNullFloat64(it.Amount),
			Modifier:        postgres.NewFloat32FromNullFloat64(it.Modifier),
			Score:           it.Score,
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

	return &domain.LogListForUserResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

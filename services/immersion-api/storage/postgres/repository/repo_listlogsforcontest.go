package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

	// Batch fetch tags for all logs
	logIDs := make([]uuid.UUID, len(entries))
	for i, it := range entries {
		logIDs[i] = it.ID
	}

	tagRows, err := r.q.ListTagsForLogs(ctx, logIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("could not fetch log tags: %w", err)
	}

	// Build a map of log_id -> tags
	tagsByLogID := make(map[uuid.UUID][]string)
	for _, row := range tagRows {
		tagsByLogID[row.LogID] = append(tagsByLogID[row.LogID], row.Tag)
	}

	res := make([]domain.Log, len(entries))
	for i, it := range entries {
		tags := tagsByLogID[it.ID]
		if tags == nil {
			tags = []string{}
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
			UnitName:        it.UnitName,
			Tags:            tags,
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

	return &domain.LogListForContestResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

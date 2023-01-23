package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ListContests(ctx context.Context, req *query.ContestListRequest) (*query.ContestListResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.ContestsMetadata(ctx, postgres.ContestsMetadataParams{
		IncludeDeleted: req.IncludeDeleted,
		UserID:         req.UserID,
		Official:       req.OfficialOnly,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not lists contests: %w", err)
	}

	contests, err := qtx.ListContests(ctx, postgres.ListContestsParams{
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
		IncludePrivate: req.IncludePrivate,
		UserID:         req.UserID,
		Official:       req.OfficialOnly,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list contests: %w", err)
	}

	res := make([]query.Contest, len(contests))
	for i, c := range contests {
		res[i] = query.Contest{
			ID:                      c.ID,
			ContestStart:            c.ContestStart,
			ContestEnd:              c.ContestEnd,
			RegistrationEnd:         c.RegistrationEnd,
			Title:                   c.Title,
			Description:             postgres.NewStringFromNullString(c.Description),
			OwnerUserID:             c.OwnerUserID,
			OwnerUserDisplayName:    c.OwnerUserDisplayName,
			Official:                c.Official,
			Private:                 c.Private,
			LanguageCodeAllowList:   c.LanguageCodeAllowList,
			ActivityTypeIDAllowList: c.ActivityTypeIDAllowList,
			CreatedAt:               c.CreatedAt,
			UpdatedAt:               c.UpdatedAt,
			Deleted:                 c.DeletedAt.Valid,
		}
	}

	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < int(meta.TotalSize) {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &query.ContestListResponse{
		Contests:      res,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}

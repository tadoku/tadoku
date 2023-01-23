package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// COMMANDS
func (r *Repository) CreateLog(ctx context.Context, req *command.LogCreateRequest) error {
	unit, err := r.q.FindUnitForTracking(ctx, postgres.FindUnitForTrackingParams{
		ID:            req.UnitID,
		LogActivityID: int16(req.ActivityID),
		LanguageCode:  postgres.NewNullString(&req.LanguageCode),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invalid unit supplied: %w", command.ErrInvalidLog)
		}
		return fmt.Errorf("could not fetch unit for tracking: %w", err)
	}

	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create log: %w", err)
	}
	qtx := r.q.WithTx(tx)

	id := uuid.New()
	if _, err = qtx.CreateLog(ctx, postgres.CreateLogParams{
		ID:                          id,
		UserID:                      req.UserID,
		LanguageCode:                req.LanguageCode,
		LogActivityID:               int16(req.ActivityID),
		UnitID:                      req.UnitID,
		Tags:                        req.Tags,
		Amount:                      req.Amount,
		Modifier:                    unit.Modifier,
		EligibleOfficialLeaderboard: req.EligibleOfficialLeaderboard,
		Description:                 postgres.NewNullString(req.Description),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create log: %w", err)
	}

	for _, registrationID := range req.RegistrationIDs {
		if err = qtx.CreateContestLogRelation(ctx, postgres.CreateContestLogRelationParams{
			RegistrationID: registrationID,
			LogID:          id,
		}); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("could not create log: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not create log: %w", err)
	}

	return nil
}

// QUERIES

func (r *Repository) FetchLogConfigurationOptions(ctx context.Context) (*query.FetchLogConfigurationOptionsResponse, error) {
	langs, err := r.q.ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	acts, err := r.q.ListActivities(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	units, err := r.q.ListUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	tags, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch log configuration options: %w", err)
	}

	options := query.FetchLogConfigurationOptionsResponse{
		Languages:  make([]query.Language, len(langs)),
		Activities: make([]query.Activity, len(acts)),
		Units:      make([]query.Unit, len(units)),
		Tags:       make([]query.Tag, len(tags)),
	}

	for i, l := range langs {
		options.Languages[i] = query.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = query.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	for i, u := range units {
		options.Units[i] = query.Unit{
			ID:            u.ID,
			LogActivityID: int(u.LogActivityID),
			Name:          u.Name,
			Modifier:      u.Modifier,
			LanguageCode:  postgres.NewStringFromNullString(u.LanguageCode),
		}
	}

	for i, t := range tags {
		options.Tags[i] = query.Tag{
			ID:            t.ID,
			LogActivityID: int(t.LogActivityID),
			Name:          t.Name,
		}
	}

	return &options, err
}

func (r *Repository) ListLogsForContestUser(ctx context.Context, req *query.LogListForContestUserRequest) (*query.LogListResponse, error) {
	_, err := r.q.FindContestById(ctx, postgres.FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	entries, err := r.q.ListLogsForContestUser(ctx, postgres.ListLogsForContestUserParams{
		ContestID:      req.ContestID,
		UserID:         req.UserID,
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &query.LogListResponse{
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	res := make([]query.Log, len(entries))
	for i, it := range entries {
		res[i] = query.Log{
			ID:           it.ID,
			UserID:       it.UserID,
			Description:  postgres.NewStringFromNullString(it.Description),
			LanguageCode: it.LanguageCode,
			LanguageName: it.LanguageName,
			ActivityID:   int(it.ActivityID),
			ActivityName: it.ActivityName,
			UnitName:     it.UnitName,
			Tags:         it.Tags,
			Amount:       it.Amount,
			Modifier:     it.Modifier,
			Score:        it.Score,
			CreatedAt:    it.CreatedAt,
			UpdatedAt:    it.UpdatedAt,
			Deleted:      it.DeletedAt.Valid,
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

	return &query.LogListResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *Repository) FindLogByID(ctx context.Context, req *query.FindLogByIDRequest) (*query.Log, error) {
	log, err := r.q.FindLogByID(ctx, postgres.FindLogByIDParams{
		IncludeDeleted: req.IncludeDeleted,
		ID:             req.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, query.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	registrations, err := r.q.FindAttachedContestRegistrationsForLog(ctx, req.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("could not fetch log details: %w", err)
	}

	refs := make([]query.ContestRegistrationReference, len(registrations))
	for i, it := range registrations {
		refs[i] = query.ContestRegistrationReference{
			RegistrationID: it.ID,
			ContestID:      it.ContestID,
			Title:          it.Title,
		}
	}

	return &query.Log{
		ID:              log.ID,
		UserID:          log.UserID,
		UserDisplayName: &log.UserDisplayName,
		Description:     postgres.NewStringFromNullString(log.Description),
		LanguageCode:    log.LanguageCode,
		LanguageName:    log.LanguageName,
		ActivityID:      int(log.ActivityID),
		ActivityName:    log.ActivityName,
		UnitName:        log.UnitName,
		Tags:            log.Tags,
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.Score,
		CreatedAt:       log.CreatedAt,
		UpdatedAt:       log.UpdatedAt,
		Deleted:         log.DeletedAt.Valid,
		Registrations:   refs,
	}, nil
}

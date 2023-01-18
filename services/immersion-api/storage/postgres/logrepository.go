package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
)

type LogRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewLogRepository(psql *sql.DB) *LogRepository {
	return &LogRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

// COMMANDS
func (r *LogRepository) CreateLog(ctx context.Context, req *logcommand.LogCreateRequest) error {
	unit, err := r.q.FindUnitForTracking(ctx, FindUnitForTrackingParams{
		ID:            req.UnitID,
		LogActivityID: int16(req.ActivityID),
		LanguageCode:  NewNullString(&req.LanguageCode),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("invalid unit supplied: %w", logcommand.ErrInvalidLog)
		}
		return fmt.Errorf("could not fetch unit for tracking: %w", err)
	}

	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create log: %w", err)
	}
	qtx := r.q.WithTx(tx)

	id := uuid.New()
	if _, err = qtx.CreateLog(ctx, CreateLogParams{
		ID:                          id,
		UserID:                      req.UserID,
		LanguageCode:                req.LanguageCode,
		LogActivityID:               int16(req.ActivityID),
		UnitID:                      req.UnitID,
		Tags:                        req.Tags,
		Amount:                      req.Amount,
		Modifier:                    unit.Modifier,
		EligibleOfficialLeaderboard: req.EligibleOfficialLeaderboard,
		Description:                 NewNullString(req.Description),
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create log: %w", err)
	}

	for _, registrationID := range req.RegistrationIDs {
		if err = qtx.CreateContestLogRelation(ctx, CreateContestLogRelationParams{
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

func (r *LogRepository) FetchLogConfigurationOptions(ctx context.Context) (*logquery.FetchLogConfigurationOptionsResponse, error) {
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

	options := logquery.FetchLogConfigurationOptionsResponse{
		Languages:  make([]logquery.Language, len(langs)),
		Activities: make([]logquery.Activity, len(acts)),
		Units:      make([]logquery.Unit, len(units)),
		Tags:       make([]logquery.Tag, len(tags)),
	}

	for i, l := range langs {
		options.Languages[i] = logquery.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	for i, a := range acts {
		options.Activities[i] = logquery.Activity{
			ID:      a.ID,
			Name:    a.Name,
			Default: a.Default,
		}
	}

	for i, u := range units {
		options.Units[i] = logquery.Unit{
			ID:            u.ID,
			LogActivityID: int(u.LogActivityID),
			Name:          u.Name,
			Modifier:      u.Modifier,
			LanguageCode:  NewStringFromNullString(u.LanguageCode),
		}
	}

	for i, t := range tags {
		options.Tags[i] = logquery.Tag{
			ID:            t.ID,
			LogActivityID: int(t.LogActivityID),
			Name:          t.Name,
		}
	}

	return &options, err
}

func (r *LogRepository) ListLogsForContestUser(ctx context.Context, req *logquery.LogListForContestUserRequest) (*logquery.LogListResponse, error) {
	_, err := r.q.FindContestById(ctx, FindContestByIdParams{ID: req.ContestID, IncludeDeleted: false})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, logquery.ErrNotFound
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	entries, err := r.q.ListLogsForContestUser(ctx, ListLogsForContestUserParams{
		ContestID:      req.ContestID,
		UserID:         req.UserID,
		StartFrom:      int32(req.Page * req.PageSize),
		PageSize:       int32(req.PageSize),
		IncludeDeleted: req.IncludeDeleted,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &logquery.LogListResponse{
				TotalSize:     0,
				NextPageToken: "",
			}, nil
		}

		return nil, fmt.Errorf("could not fetch logs list: %w", err)
	}

	res := make([]logquery.Log, len(entries))
	for i, it := range entries {
		res[i] = logquery.Log{
			ID:           it.ID,
			UserID:       it.UserID,
			Description:  NewStringFromNullString(it.Description),
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

	return &logquery.LogListResponse{
		Logs:          res,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *LogRepository) FindLogByID(ctx context.Context, req *logquery.FindLogByIDRequest) (*logquery.Log, error) {
	return nil, nil
}

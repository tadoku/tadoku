package contests

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type ContestsRepository struct {
	db *pgxpool.Pool
}

func NewContestsRepository(db *pgxpool.Pool) *ContestsRepository {
	return &ContestsRepository{db: db}
}

func (r *ContestsRepository) CountContestsCreatedByUserForYear(ctx context.Context, userID uuid.UUID, year int32) (int64, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}

	count, err := queries.New(executor).CountContestsCreatedByUserForYear(ctx, queries.CountContestsCreatedByUserForYearParams{
		OwnerUserID: pgtype.UUID{Bytes: userID, Valid: true},
		Year:        year,
	})
	if err != nil {
		return 0, fmt.Errorf("count contests created by user for year: %w", err)
	}
	return count, nil
}

func (r *ContestsRepository) ListContests(ctx context.Context, parameters ListParameters) ([]Contest, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListContests(ctx, queries.ListContestsParams{
		IncludeDeleted: parameters.IncludeDeleted,
		UserID:         nullableUUID(parameters.UserID),
		Official:       parameters.Official,
		IncludePrivate: parameters.IncludePrivate(),
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list contests: %w", err)
	}

	result := make([]Contest, 0, len(rows))
	var total int
	for _, row := range rows {
		total = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}
		result = append(result, Contest{
			ID:                      uuid.UUID(row.ID.Bytes),
			ContestStart:            row.ContestStart.Time,
			ContestEnd:              row.ContestEnd.Time,
			RegistrationEnd:         row.RegistrationEnd.Time,
			Title:                   row.Title.String,
			Description:             nullableString(row.Description),
			OwnerUserID:             uuid.UUID(row.OwnerUserID.Bytes),
			OwnerUserDisplayName:    row.OwnerUserDisplayName.String,
			Official:                row.Official.Bool,
			Private:                 row.Private.Bool,
			LanguageCodeAllowList:   row.LanguageCodeAllowList,
			ActivityTypeIDAllowList: row.ActivityTypeIDAllowList,
			CreatedAt:               row.CreatedAt.Time,
			UpdatedAt:               row.UpdatedAt.Time,
			Deleted:                 row.DeletedAt.Valid,
		})
	}
	return result, total, nil
}

func (r *ContestsRepository) FindContestByID(ctx context.Context, parameters FindParameters) (*Contest, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindContestByID(ctx, queries.FindContestByIDParams{
		ID:             pgtype.UUID{Bytes: parameters.ID, Valid: true},
		IncludeDeleted: parameters.IncludeDeleted(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find contest by ID: %w", err)
	}

	result := contestFromRow(row)
	return &result, nil
}

func (r *ContestsRepository) FindLatestOfficialContest(ctx context.Context) (*Contest, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindLatestOfficialContest(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find latest official contest: %w", err)
	}

	result := contestFromRow(queries.FindContestByIDRow(row))
	return &result, nil
}

func (r *ContestsRepository) ListLanguagesForContest(ctx context.Context, contestID uuid.UUID) ([]Language, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLanguagesForContest(ctx, pgtype.UUID{Bytes: contestID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list contest languages: %w", err)
	}
	result := make([]Language, 0, len(rows))
	for _, row := range rows {
		result = append(result, Language{Code: row.Code, Name: row.Name})
	}
	return result, nil
}

func (r *ContestsRepository) ListLanguages(ctx context.Context) ([]Language, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLanguages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	result := make([]Language, 0, len(rows))
	for _, row := range rows {
		result = append(result, Language{Code: row.Code, Name: row.Name})
	}
	return result, nil
}

func contestFromRow(row queries.FindContestByIDRow) Contest {
	return Contest{
		ID:                      uuid.UUID(row.ID.Bytes),
		ContestStart:            row.ContestStart.Time,
		ContestEnd:              row.ContestEnd.Time,
		RegistrationEnd:         row.RegistrationEnd.Time,
		Title:                   row.Title,
		Description:             nullableString(row.Description),
		OwnerUserID:             uuid.UUID(row.OwnerUserID.Bytes),
		OwnerUserDisplayName:    row.OwnerUserDisplayName,
		Official:                row.Official,
		Private:                 row.Private,
		LanguageCodeAllowList:   row.LanguageCodeAllowList,
		ActivityTypeIDAllowList: row.ActivityTypeIDAllowList,
		CreatedAt:               row.CreatedAt.Time,
		UpdatedAt:               row.UpdatedAt.Time,
		Deleted:                 row.DeletedAt.Valid,
	}
}

func nullableUUID(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *value, Valid: true}
}

func nullableString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

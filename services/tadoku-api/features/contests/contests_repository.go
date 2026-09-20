package contests

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *ContestsRepository) UpsertContestCreator(ctx context.Context, parameters CreateParameters, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	_, err = queries.New(executor).UpsertContestCreator(ctx, queries.UpsertContestCreatorParams{
		ID:               pgtype.UUID{Bytes: parameters.OwnerUserID(), Valid: true},
		DisplayName:      parameters.OwnerUserDisplayName(),
		SessionCreatedAt: pgtype.Timestamp{Time: parameters.SessionCreatedAt(), Valid: true},
		CreatedAt:        pgtype.Timestamp{Time: now, Valid: true},
		UpdatedAt:        pgtype.Timestamp{Time: now, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAccountDeletionInProgress
	}
	if err != nil {
		return fmt.Errorf("upsert contest creator: %w", err)
	}
	return nil
}

func (r *ContestsRepository) LockContestCreator(ctx context.Context, userID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	creator, err := queries.New(executor).LockContestCreator(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrContestCreatorNotFound
	}
	if err != nil {
		return fmt.Errorf("lock contest creator: %w", err)
	}
	if creator.DeletionLockedAt.Valid || creator.DeletedAt.Valid {
		return ErrAccountDeletionInProgress
	}
	return nil
}

func (r *ContestsRepository) LanguagesExist(ctx context.Context, codes []string) (bool, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return false, err
	}
	exists, err := queries.New(executor).LanguagesExist(ctx, codes)
	if err != nil {
		return false, fmt.Errorf("check contest languages: %w", err)
	}
	return exists, nil
}

func (r *ContestsRepository) CreateContest(ctx context.Context, parameters CreateParameters) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	err = queries.New(executor).CreateContest(ctx, queries.CreateContestParams{
		ID:                      pgtype.UUID{Bytes: parameters.ID(), Valid: true},
		OwnerUserID:             pgtype.UUID{Bytes: parameters.OwnerUserID(), Valid: true},
		OwnerUserDisplayName:    parameters.OwnerUserDisplayName(),
		Official:                parameters.Official,
		Private:                 parameters.Private,
		ContestStart:            pgtype.Date{Time: parameters.ContestStart, Valid: true},
		ContestEnd:              pgtype.Date{Time: parameters.ContestEnd, Valid: true},
		RegistrationEnd:         pgtype.Date{Time: parameters.RegistrationEnd, Valid: true},
		Title:                   parameters.Title,
		Description:             text(parameters.Description),
		LanguageCodeAllowList:   parameters.LanguageCodeAllowList,
		ActivityTypeIDAllowList: parameters.ActivityTypeIDAllowList,
		CreatedAt:               pgtype.Timestamp{Time: parameters.CreatedAt(), Valid: true},
		UpdatedAt:               pgtype.Timestamp{Time: parameters.UpdatedAt(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create contest: %w", err)
	}
	return nil
}

func (r *ContestsRepository) FindCreatedContestByID(ctx context.Context, id uuid.UUID) (*Contest, error) {
	return r.FindContestByID(ctx, FindParameters{ID: id})
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

func text(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

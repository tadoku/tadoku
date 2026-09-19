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

func (r *ContestsRepository) ListContests(ctx context.Context, parameters ListParameters) ([]Contest, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	userID := nullableUUID(parameters.UserID)
	q := queries.New(executor)
	total, err := q.ContestsMetadata(ctx, queries.ContestsMetadataParams{
		IncludeDeleted: parameters.IncludeDeleted,
		UserID:         userID,
		Official:       parameters.Official,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count contests: %w", err)
	}
	rows, err := q.ListContests(ctx, queries.ListContestsParams{
		IncludeDeleted: parameters.IncludeDeleted,
		UserID:         userID,
		Official:       parameters.Official,
		IncludePrivate: parameters.IncludePrivate(),
		StartFrom:      int32(parameters.Page * parameters.PageSize),
		PageSize:       int32(parameters.PageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list contests: %w", err)
	}

	result := make([]Contest, 0, len(rows))
	for _, row := range rows {
		result = append(result, Contest{
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
		})
	}
	return result, int(total), nil
}

func (r *ContestsRepository) FindContestByID(ctx context.Context, parameters FindParameters) (*ContestView, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	q := queries.New(executor)
	row, err := q.FindContestByID(ctx, queries.FindContestByIDParams{
		ID:             pgtype.UUID{Bytes: parameters.ID, Valid: true},
		IncludeDeleted: parameters.IncludeDeleted(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find contest by ID: %w", err)
	}

	languages, err := listContestLanguages(ctx, q, row.ID, row.LanguageCodeAllowList)
	if err != nil {
		return nil, err
	}
	return contestView(row, languages), nil
}

func (r *ContestsRepository) FindLatestOfficialContest(ctx context.Context) (*ContestView, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	q := queries.New(executor)
	row, err := q.FindLatestOfficialContest(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrContestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find latest official contest: %w", err)
	}

	languages, err := listContestLanguages(ctx, q, row.ID, row.LanguageCodeAllowList)
	if err != nil {
		return nil, err
	}
	return contestView(queries.FindContestByIDRow(row), languages), nil
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

func listContestLanguages(ctx context.Context, q *queries.Queries, contestID pgtype.UUID, allowList []string) ([]Language, error) {
	if len(allowList) == 0 {
		return nil, nil
	}
	rows, err := q.ListLanguagesForContest(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("list contest languages: %w", err)
	}
	result := make([]Language, 0, len(rows))
	for _, row := range rows {
		result = append(result, Language{Code: row.Code, Name: row.Name})
	}
	return result, nil
}

func contestView(row queries.FindContestByIDRow, languages []Language) *ContestView {
	return &ContestView{
		ID:                   uuid.UUID(row.ID.Bytes),
		ContestStart:         row.ContestStart.Time,
		ContestEnd:           row.ContestEnd.Time,
		RegistrationEnd:      row.RegistrationEnd.Time,
		Title:                row.Title,
		Description:          nullableString(row.Description),
		OwnerUserID:          uuid.UUID(row.OwnerUserID.Bytes),
		OwnerUserDisplayName: row.OwnerUserDisplayName,
		Official:             row.Official,
		Private:              row.Private,
		AllowedLanguages:     languages,
		AllowedActivities:    make([]Activity, 0, len(row.ActivityTypeIDAllowList)),
		allowedActivityIDs:   row.ActivityTypeIDAllowList,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		Deleted:              row.DeletedAt.Valid,
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

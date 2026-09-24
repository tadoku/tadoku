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
	"github.com/tadoku/tadoku/services/tadoku-api/domain/leaderboardoutbox"
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
		OwnerUserID: postgres.UUID(userID),
		Year:        year,
	})
	if err != nil {
		return 0, fmt.Errorf("count contests created by user for year: %w", err)
	}
	return count, nil
}

func (r *ContestsRepository) FindRegistrationForUser(ctx context.Context, userID, contestID uuid.UUID) (*Registration, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindContestRegistrationForUser(ctx, queries.FindContestRegistrationForUserParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find contest registration: %w", err)
	}

	return &Registration{
		ID:              uuid.UUID(row.ID.Bytes),
		ContestID:       uuid.UUID(row.ContestID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		UserDisplayName: row.UserDisplayName,
		LanguageCodes:   row.LanguageCodes,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *ContestsRepository) FindRegistrationWithContestForUser(ctx context.Context, userID, contestID uuid.UUID) (*Registration, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindContestRegistrationWithContestForUser(ctx, queries.FindContestRegistrationWithContestForUserParams{
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find contest registration with contest: %w", err)
	}

	return &Registration{
		ID:              uuid.UUID(row.ID.Bytes),
		ContestID:       uuid.UUID(row.ContestID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		UserDisplayName: row.UserDisplayName,
		LanguageCodes:   row.LanguageCodes,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		Contest: &ContestView{
			ID:                uuid.UUID(row.ContestID.Bytes),
			ContestStart:      row.ContestStart.Time,
			ContestEnd:        row.ContestEnd.Time,
			RegistrationEnd:   row.RegistrationEnd.Time,
			Title:             row.Title,
			Description:       postgres.TextPointer(row.Description),
			Private:           row.Private,
			Official:          row.Official,
			AllowedLanguages:  []Language{},
			AllowedActivities: []Activity{},
		},
	}, nil
}

func (r *ContestsRepository) ListOngoingRegistrations(ctx context.Context, userID uuid.UUID, now time.Time) ([]Registration, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListOngoingContestRegistrations(ctx, queries.ListOngoingContestRegistrationsParams{
		UserID: postgres.UUID(userID),
		Now:    postgres.Timestamp(now),
	})
	if err != nil {
		return nil, fmt.Errorf("list ongoing contest registrations: %w", err)
	}

	result := make([]Registration, 0, len(rows))
	for _, row := range rows {
		registration := registrationFromRow(queries.ListYearlyContestRegistrationsRow{
			ID:                      row.ID,
			ContestID:               row.ContestID,
			UserID:                  row.UserID,
			LanguageCodes:           row.LanguageCodes,
			UserDisplayName:         row.UserDisplayName,
			ActivityTypeIDAllowList: row.ActivityTypeIDAllowList,
			RegistrationEnd:         row.RegistrationEnd,
			ContestStart:            row.ContestStart,
			ContestEnd:              row.ContestEnd,
			Private:                 row.Private,
			Official:                row.Official,
			Title:                   row.Title,
			Description:             row.Description,
		})
		registration.Contest.OwnerUserID = uuid.UUID(row.OwnerUserID.Bytes)
		registration.Contest.OwnerUserDisplayName = row.OwnerUserDisplayName

		result = append(result, registration)
	}

	return result, nil
}

func (r *ContestsRepository) ListYearlyRegistrations(ctx context.Context, userID uuid.UUID, year int32, includePrivate bool) ([]Registration, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListYearlyContestRegistrations(ctx, queries.ListYearlyContestRegistrationsParams{
		UserID:         postgres.UUID(userID),
		Year:           year,
		IncludePrivate: includePrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("list yearly contest registrations: %w", err)
	}

	result := make([]Registration, 0, len(rows))
	for _, row := range rows {
		result = append(result, registrationFromRow(row))
	}

	return result, nil
}

func (r *ContestsRepository) DetachContestLogsForLanguages(
	ctx context.Context,
	userID uuid.UUID,
	contestID uuid.UUID,
	languageCodes []string,
) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).DetachContestLogsForLanguages(ctx, queries.DetachContestLogsForLanguagesParams{
		ContestID:     postgres.UUID(contestID),
		UserID:        postgres.UUID(userID),
		LanguageCodes: languageCodes,
	})
	if err != nil {
		return fmt.Errorf("detach contest logs for languages: %w", err)
	}

	return nil
}

func (r *ContestsRepository) UpsertRegistration(ctx context.Context, registration Registration) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).UpsertContestRegistration(ctx, queries.UpsertContestRegistrationParams{
		ID:            postgres.UUID(registration.ID),
		ContestID:     postgres.UUID(registration.ContestID),
		UserID:        postgres.UUID(registration.UserID),
		LanguageCodes: registration.LanguageCodes,
		CreatedAt:     postgres.Timestamp(registration.CreatedAt),
		UpdatedAt:     postgres.Timestamp(registration.UpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("upsert contest registration: %w", err)
	}

	return nil
}

func (r *ContestsRepository) InsertContestScoreRefresh(ctx context.Context, userID, contestID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	if err := queries.New(executor).InsertContestScoreRefresh(ctx, queries.InsertContestScoreRefreshParams{
		EventType: string(leaderboardoutbox.RefreshContestScore),
		UserID:    postgres.UUID(userID),
		ContestID: postgres.UUID(contestID),
	}); err != nil {
		return fmt.Errorf("insert contest score refresh: %w", err)
	}

	return nil
}

func (r *ContestsRepository) InsertOfficialScoresRefresh(ctx context.Context, userID uuid.UUID, year int16) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	if err := queries.New(executor).InsertOfficialScoresRefresh(ctx, queries.InsertOfficialScoresRefreshParams{
		EventType: string(leaderboardoutbox.RefreshOfficialScores),
		UserID:    postgres.UUID(userID),
		Year:      pgtype.Int2{Int16: year, Valid: true},
	}); err != nil {
		return fmt.Errorf("insert official scores refresh: %w", err)
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

func (r *ContestsRepository) CreateContest(ctx context.Context, contest Contest) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	err = queries.New(executor).CreateContest(ctx, queries.CreateContestParams{
		ID:                      postgres.UUID(contest.ID),
		OwnerUserID:             postgres.UUID(contest.OwnerUserID),
		OwnerUserDisplayName:    contest.OwnerUserDisplayName,
		Official:                contest.Official,
		Private:                 contest.Private,
		ContestStart:            pgtype.Date{Time: contest.ContestStart, Valid: true},
		ContestEnd:              pgtype.Date{Time: contest.ContestEnd, Valid: true},
		RegistrationEnd:         pgtype.Date{Time: contest.RegistrationEnd, Valid: true},
		Title:                   contest.Title,
		Description:             postgres.NullableText(contest.Description),
		LanguageCodeAllowList:   contest.LanguageCodeAllowList,
		ActivityTypeIDAllowList: contest.ActivityTypeIDAllowList,
		CreatedAt:               postgres.Timestamp(contest.CreatedAt),
		UpdatedAt:               postgres.Timestamp(contest.UpdatedAt),
	})
	if err != nil {
		return fmt.Errorf("create contest: %w", err)
	}
	return nil
}

func (r *ContestsRepository) ListContests(ctx context.Context, parameters ListParameters) ([]Contest, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListContests(ctx, queries.ListContestsParams{
		IncludeDeleted: parameters.IncludeDeleted,
		UserID:         postgres.NullableUUID(parameters.UserID),
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
			Description:             postgres.TextPointer(row.Description),
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
		ID:             postgres.UUID(parameters.ID),
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

func (r *ContestsRepository) FetchContestSummary(ctx context.Context, contestID uuid.UUID) (*ContestSummary, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FetchContestSummary(ctx, postgres.UUID(contestID))
	if err != nil {
		return nil, fmt.Errorf("fetch contest summary: %w", err)
	}

	return &ContestSummary{
		ParticipantCount: int(row.ParticipantCount),
		LanguageCount:    int(row.LanguageCount),
		TotalScore:       row.TotalScore,
	}, nil
}

func (r *ContestsRepository) ListLanguagesForContest(ctx context.Context, contestID uuid.UUID) ([]Language, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListLanguagesForContest(ctx, postgres.UUID(contestID))
	if err != nil {
		return nil, fmt.Errorf("list contest languages: %w", err)
	}
	result := make([]Language, 0, len(rows))
	for _, row := range rows {
		result = append(result, Language{Code: row.Code, Name: row.Name})
	}
	return result, nil
}

func registrationFromRow(row queries.ListYearlyContestRegistrationsRow) Registration {
	return Registration{
		ID:              uuid.UUID(row.ID.Bytes),
		ContestID:       uuid.UUID(row.ContestID.Bytes),
		UserID:          uuid.UUID(row.UserID.Bytes),
		UserDisplayName: row.UserDisplayName,
		LanguageCodes:   row.LanguageCodes,
		Contest: &ContestView{
			ID:                 uuid.UUID(row.ContestID.Bytes),
			ContestStart:       row.ContestStart.Time,
			ContestEnd:         row.ContestEnd.Time,
			RegistrationEnd:    row.RegistrationEnd.Time,
			Title:              row.Title,
			Description:        postgres.TextPointer(row.Description),
			Official:           row.Official,
			Private:            row.Private,
			AllowedLanguages:   []Language{},
			AllowedActivities:  make([]Activity, 0, len(row.ActivityTypeIDAllowList)),
			allowedActivityIDs: row.ActivityTypeIDAllowList,
		},
	}
}

func contestFromRow(row queries.FindContestByIDRow) Contest {
	return Contest{
		ID:                      uuid.UUID(row.ID.Bytes),
		ContestStart:            row.ContestStart.Time,
		ContestEnd:              row.ContestEnd.Time,
		RegistrationEnd:         row.RegistrationEnd.Time,
		Title:                   row.Title,
		Description:             postgres.TextPointer(row.Description),
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

package leaderboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) contestExists(ctx context.Context, id uuid.UUID) (bool, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return false, err
	}
	exists, err := q.ContestExists(ctx, uuidValue(id))
	if err != nil {
		return false, fmt.Errorf("check contest: %w", err)
	}
	return exists, nil
}

func (r *Repository) contest(ctx context.Context, request ContestRequest) (*Leaderboard, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.LeaderboardForContest(ctx, queries.LeaderboardForContestParams{
		ContestID:    uuidValue(request.ContestID),
		LanguageCode: textValue(request.LanguageCode),
		ActivityID:   int32Value(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch contest leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) yearly(ctx context.Context, request YearlyRequest) (*Leaderboard, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.YearlyLeaderboard(ctx, queries.YearlyLeaderboardParams{
		Year:         int16(request.Year),
		LanguageCode: textValue(request.LanguageCode),
		ActivityID:   int32Value(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch yearly leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) global(ctx context.Context, request Request) (*Leaderboard, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.GlobalLeaderboard(ctx, queries.GlobalLeaderboardParams{
		LanguageCode: textValue(request.LanguageCode),
		ActivityID:   int32Value(request.ActivityID),
		StartFrom:    int32(request.Page * request.PageSize),
		PageSize:     int32(request.PageSize),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch global leaderboard: %w", err)
	}
	entries := make([]Entry, len(rows))
	for i, row := range rows {
		entries[i] = Entry{
			Rank:            int(row.Rank),
			UserID:          uuid.UUID(row.UserID.Bytes),
			UserDisplayName: row.UserDisplayName,
			Score:           row.Score,
			IsTie:           row.IsTie,
		}
	}
	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalSize)
	}
	return result(entries, total, request.Page, request.PageSize), nil
}

func (r *Repository) allContestScores(ctx context.Context, id uuid.UUID) ([]score, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.ContestLeaderboardAllScores(ctx, uuidValue(id))
	if err != nil {
		return nil, fmt.Errorf("fetch all contest leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func (r *Repository) allYearlyScores(ctx context.Context, year int) ([]score, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.YearlyLeaderboardAllScores(ctx, int16(year))
	if err != nil {
		return nil, fmt.Errorf("fetch all yearly leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func (r *Repository) allGlobalScores(ctx context.Context) ([]score, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.GlobalLeaderboardAllScores(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch all global leaderboard scores: %w", err)
	}
	res := make([]score, len(rows))
	for i, row := range rows {
		res[i] = score{userID: uuid.UUID(row.UserID.Bytes), value: float64(row.Score)}
	}
	return res, nil
}

func (r *Repository) displayNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	values := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		values[i] = uuidValue(id)
	}
	rows, err := q.FindUserDisplayNames(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("fetch user display names: %w", err)
	}
	names := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		names[uuid.UUID(row.ID.Bytes)] = row.DisplayName
	}
	return names, nil
}

func (r *Repository) queries(ctx context.Context) (*queries.Queries, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return queries.New(executor), nil
}

func uuidValue(value uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: value, Valid: true} }
func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}
func int32Value(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func result(entries []Entry, total, currentPage, pageSize int) *Leaderboard {
	next := ""
	if currentPage*pageSize+pageSize < total {
		next = fmt.Sprint(currentPage + 1)
	}
	return &Leaderboard{
		Entries:       entries,
		TotalSize:     total,
		NextPageToken: next,
	}
}

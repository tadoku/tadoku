package tokens

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) lockTenant(ctx context.Context) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	if _, ok := db.(pgx.Tx); !ok {
		return errors.New("access token creation requires a tenant transaction")
	}
	_, err = queries.New(db).LockTenant(ctx)
	return err
}

func (r *Repository) insert(ctx context.Context, row Token, hash string, activeLimit int) (bool, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return false, err
	}

	scopes := make([]string, 0, len(row.Scopes))
	for _, scope := range row.Scopes {
		scopes = append(scopes, string(scope))
	}

	_, err = queries.New(db).Insert(ctx, queries.InsertParams{
		ID:          postgres.UUID(row.ID),
		Name:        row.Name,
		TokenHash:   hash,
		Scopes:      scopes,
		CreatedAt:   postgres.Timestamptz(row.CreatedAt),
		ExpiresAt:   postgres.Timestamptz(row.ExpiresAt),
		ActiveLimit: int64(activeLimit),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) verify(ctx context.Context, hash string, now time.Time) (Token, bool, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return Token{}, false, err
	}

	row, err := queries.New(db).Verify(ctx, queries.VerifyParams{TokenHash: hash, Now: postgres.Timestamptz(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return Token{}, false, nil
	}
	if err != nil {
		return Token{}, false, err
	}
	return tokenFromRow(queries.ListRow(row)), true, nil
}

func (r *Repository) list(ctx context.Context) ([]Token, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(db).List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Token, 0, len(rows))
	for _, row := range rows {
		result = append(result, tokenFromRow(row))
	}
	return result, nil
}

func (r *Repository) revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	_, err = queries.New(db).Revoke(ctx, queries.RevokeParams{ID: postgres.UUID(id), Now: postgres.Timestamptz(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return errx.NewNotFoundError("access token not found")
	}
	return err
}

func tokenFromRow(row queries.ListRow) Token {
	scopes := make([]Scope, 0, len(row.Scopes))
	for _, scope := range row.Scopes {
		scopes = append(scopes, Scope(scope))
	}
	result := Token{
		ID:        row.ID.Bytes,
		Name:      row.Name,
		Scopes:    scopes,
		CreatedAt: row.CreatedAt.Time,
		ExpiresAt: row.ExpiresAt.Time,
	}
	if row.LastUsedAt.Valid {
		result.LastUsedAt = &row.LastUsedAt.Time
	}
	if row.RevokedAt.Valid {
		result.RevokedAt = &row.RevokedAt.Time
	}
	return result
}

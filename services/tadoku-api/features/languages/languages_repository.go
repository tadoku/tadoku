package languages

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type LanguagesRepository struct {
	db *pgxpool.Pool
}

func NewLanguagesRepository(db *pgxpool.Pool) *LanguagesRepository {
	return &LanguagesRepository{db: db}
}

func (r *LanguagesRepository) CreateLanguage(ctx context.Context, parameters CreateLanguageParameters) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).CreateLanguage(ctx, queries.CreateLanguageParams{
		Code: parameters.Code,
		Name: parameters.Name,
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrLanguageAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create language: %w", err)
	}
	return nil
}

func (r *LanguagesRepository) ListLanguages(ctx context.Context) ([]Language, error) {
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
		result = append(result, Language{
			Code: row.Code,
			Name: row.Name,
		})
	}
	return result, nil
}

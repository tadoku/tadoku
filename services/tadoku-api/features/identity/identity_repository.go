package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type credentials struct {
	account      Account
	passwordHash string
	active       bool
}

func (r *Repository) register(ctx context.Context, result LoginResult, passwordHash string) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}

	account := result.Session.Account
	err = queries.New(db).RegisterAccount(ctx, queries.RegisterAccountParams{
		UserID:          postgres.UUID(account.ID),
		DisplayName:     account.DisplayName.value,
		Email:           account.Email.value,
		EmailNormalized: account.Email.value,
		PasswordHash:    passwordHash,
		Now:             postgres.Timestamp(account.CreatedAt),
		TokenHash:       hashToken(result.Cookie.Value),
		AuthenticatedAt: postgres.Timestamptz(result.Session.AuthenticatedAt),
		ExpiresAt:       postgres.Timestamptz(result.Session.ExpiresAt),
	})
	if postgres.IsUniqueViolation(err) {
		return ErrEmailTaken
	}
	return err
}

func (r *Repository) credentials(ctx context.Context, email Email) (credentials, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return credentials{}, err
	}
	row, err := queries.New(db).FindCredentials(ctx, email.value)
	if errors.Is(err, pgx.ErrNoRows) {
		return credentials{}, errAccountNotFound
	}
	if err != nil {
		return credentials{}, err
	}
	return credentials{
		account: Account{
			ID:          uuid.UUID(row.UserID.Bytes),
			Email:       Email{row.Email},
			DisplayName: DisplayName{row.DisplayName},
			CreatedAt:   row.CreatedAt.Time,
		},
		passwordHash: row.PasswordHash,
		active:       row.Active,
	}, nil
}

func (r *Repository) createSession(
	ctx context.Context,
	session AuthenticatedSession,
	tokenHash, passwordHash string,
) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	count, err := queries.New(db).CreateSession(ctx, queries.CreateSessionParams{
		PasswordHash:    passwordHash,
		UserID:          postgres.UUID(session.Account.ID),
		TokenHash:       tokenHash,
		AuthenticatedAt: postgres.Timestamptz(session.AuthenticatedAt),
		ExpiresAt:       postgres.Timestamptz(session.ExpiresAt),
	})
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrWrongPassword
	}
	return nil
}

func (r *Repository) session(ctx context.Context, tokenHash string, now time.Time) (Session, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	row, err := queries.New(db).FindSession(ctx, queries.FindSessionParams{
		TokenHash: tokenHash,
		Now:       postgres.Timestamptz(now),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return GuestSession{}, nil
	}
	if err != nil {
		return nil, err
	}
	return AuthenticatedSession{
		Account: Account{
			ID:          uuid.UUID(row.UserID.Bytes),
			Email:       Email{row.Email},
			DisplayName: DisplayName{row.DisplayName},
			CreatedAt:   row.CreatedAt.Time,
		},
		AuthenticatedAt: row.AuthenticatedAt.Time,
		ExpiresAt:       row.ExpiresAt.Time,
	}, nil
}

func (r *Repository) revokeSession(ctx context.Context, tokenHash string) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).DeleteSession(ctx, tokenHash)
}

func (r *Repository) developmentAccountExists(ctx context.Context, id uuid.UUID) (bool, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return false, err
	}
	return queries.New(db).DevelopmentAccountExists(ctx, postgres.UUID(id))
}

func (r *Repository) seedDevelopmentUser(ctx context.Context, id uuid.UUID, name DisplayName) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).InsertDevelopmentUser(ctx, queries.InsertDevelopmentUserParams{
		ID:          postgres.UUID(id),
		DisplayName: name.value,
	})
}

func (r *Repository) seedDevelopmentAccount(
	ctx context.Context,
	id uuid.UUID,
	email Email,
	passwordHash string,
) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).InsertDevelopmentAccount(ctx, queries.InsertDevelopmentAccountParams{
		UserID:       postgres.UUID(id),
		Email:        email.value,
		PasswordHash: passwordHash,
	})
}

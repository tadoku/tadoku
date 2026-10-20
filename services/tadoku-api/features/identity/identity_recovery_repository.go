package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *Repository) lockRecoveryAccount(ctx context.Context, email Email) (credentials, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return credentials{}, err
	}
	row, err := queries.New(db).LockRecoveryAccount(ctx, email.value)
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
		active:       true,
	}, nil
}

func (r *Repository) lockAccount(ctx context.Context, id uuid.UUID) (credentials, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return credentials{}, err
	}
	row, err := queries.New(db).LockAccount(ctx, postgres.UUID(id))
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
		active:       true,
	}, nil
}

func (r *Repository) createRecoveryToken(ctx context.Context, id uuid.UUID, hash string, expires time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).CreateRecoveryToken(ctx, queries.CreateRecoveryTokenParams{
		UserID:    postgres.UUID(id),
		TokenHash: hash,
		ExpiresAt: postgres.Timestamptz(expires),
	})
}

func (r *Repository) recoveryAccount(ctx context.Context, hash string) (uuid.UUID, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := queries.New(db).RecoveryAccount(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrInvalidRecoveryToken
	}
	return uuid.UUID(id.Bytes), err
}

func (r *Repository) useRecoveryToken(ctx context.Context, hash string, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	count, err := queries.New(db).UseRecoveryToken(ctx, queries.UseRecoveryTokenParams{
		TokenHash: hash,
		Now:       postgres.Timestamptz(now),
	})
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrInvalidRecoveryToken
	}
	return nil
}

func (r *Repository) invalidateRecoveryTokens(ctx context.Context, id uuid.UUID, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).InvalidateRecoveryTokens(ctx, queries.InvalidateRecoveryTokensParams{
		UserID: postgres.UUID(id),
		Now:    postgres.Timestamptz(now),
	})
}

func (r *Repository) revokeOtherSessions(ctx context.Context, id uuid.UUID, keepHash string) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).RevokeOtherSessions(ctx, queries.RevokeOtherSessionsParams{
		UserID:        postgres.UUID(id),
		KeepTokenHash: keepHash,
	})
}

func (r *Repository) updateEmail(ctx context.Context, id uuid.UUID, email Email, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	err = queries.New(db).UpdateAccountEmail(ctx, queries.UpdateAccountEmailParams{
		UserID: postgres.UUID(id),
		Email:  email.value,
		Now:    postgres.Timestamptz(now),
	})
	if postgres.IsUniqueViolation(err) {
		return ErrEmailTaken
	}
	return err
}

func (r *Repository) updateDisplayName(ctx context.Context, id uuid.UUID, name DisplayName, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).UpdateDisplayName(ctx, queries.UpdateDisplayNameParams{
		UserID:      postgres.UUID(id),
		DisplayName: name.value,
		Now:         postgres.Timestamp(now),
	})
}

func (r *Repository) updatePassword(ctx context.Context, id uuid.UUID, hash string, now time.Time) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).UpdateAccountPassword(ctx, queries.UpdateAccountPasswordParams{
		UserID:       postgres.UUID(id),
		PasswordHash: hash,
		Now:          postgres.Timestamptz(now),
	})
}

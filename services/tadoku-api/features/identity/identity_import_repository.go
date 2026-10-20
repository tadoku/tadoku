package identity

import (
	"context"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *Repository) importUsers(ctx context.Context) (map[uuid.UUID]targetUser, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(db).ImportUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make(map[uuid.UUID]targetUser, len(rows))
	for _, row := range rows {
		id := uuid.UUID(row.ID.Bytes)
		users[id] = targetUser{displayName: row.DisplayName, deleted: row.Deleted}
	}
	return users, nil
}

func (r *Repository) importAccounts(ctx context.Context) (map[uuid.UUID]sourceIdentity, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(db).ImportAccounts(ctx)
	if err != nil {
		return nil, err
	}

	accounts := make(map[uuid.UUID]sourceIdentity, len(rows))
	for _, row := range rows {
		id := uuid.UUID(row.UserID.Bytes)
		accounts[id] = sourceIdentity{
			id:              id,
			email:           row.Email,
			normalizedEmail: row.EmailNormalized,
			hash:            row.PasswordHash,
			state:           row.State,
			createdAt:       row.CreatedAt.Time,
		}
	}
	return accounts, nil
}

func (r *Repository) insertImportedUser(ctx context.Context, record sourceIdentity) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).InsertImportedUser(ctx, queries.InsertImportedUserParams{
		ID:          postgres.UUID(record.id),
		DisplayName: record.displayName,
		CreatedAt:   postgres.Timestamp(record.createdAt),
		UpdatedAt:   postgres.Timestamp(record.updatedAt),
	})
}

func (r *Repository) updateImportedUserName(ctx context.Context, record sourceIdentity) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).UpdateImportedUserName(ctx, queries.UpdateImportedUserNameParams{
		ID:          postgres.UUID(record.id),
		DisplayName: record.displayName,
	})
}

func (r *Repository) insertImportedAccount(ctx context.Context, record sourceIdentity) error {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	return queries.New(db).InsertImportedAccount(ctx, queries.InsertImportedAccountParams{
		UserID:          postgres.UUID(record.id),
		Email:           record.email,
		EmailNormalized: record.normalizedEmail,
		PasswordHash:    record.hash,
		State:           record.state,
		CreatedAt:       postgres.Timestamptz(record.createdAt),
		UpdatedAt:       postgres.Timestamptz(record.updatedAt),
	})
}

func (r *Repository) importAccountCount(ctx context.Context) (int64, error) {
	db, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return 0, err
	}
	return queries.New(db).ImportAccountCount(ctx)
}

package identity

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type targetUser struct {
	displayName string
	deleted     bool
}

// Use only for an explicitly authorized import with distinct source and target pools.
func ImportKratosAccounts(
	ctx context.Context,
	source, target *pgxpool.Pool,
	expect int,
	apply bool,
	output io.Writer,
) error {
	records, err := readSource(ctx, source, expect, output)
	if err != nil {
		return err
	}
	return importTarget(ctx, target, records, apply, output)
}

func importTarget(
	ctx context.Context,
	pool *pgxpool.Pool,
	records []sourceIdentity,
	apply bool,
	output io.Writer,
) error {
	ctx = tenant.WithKey(ctx, tenant.Production())
	newUsers, newAccounts, updatedUsers := 0, 0, 0
	repository := NewRepository(pool)
	var operationErr error
	err := postgres.RunInTransaction(ctx, pool, func(ctx context.Context) (err error) {
		defer func() { operationErr = err }()
		db, err := postgres.Executor(ctx, pool)
		if err != nil {
			return databaseError("target transaction", err)
		}
		if _, err := db.Exec(ctx, "select set_config('lock_timeout', '5s', true)"); err != nil {
			return databaseError("target lock", err)
		}
		if _, err := db.Exec(ctx, "lock table users, accounts in share row exclusive mode"); err != nil {
			return databaseError("target lock", err)
		}

		users, err := repository.importUsers(ctx)
		if err != nil {
			return databaseError("target read", err)
		}
		accounts, err := repository.importAccounts(ctx)
		if err != nil {
			return databaseError("target read", err)
		}
		sourceIDs := make(map[uuid.UUID]bool, len(records))
		for _, record := range records {
			sourceIDs[record.id] = true
		}
		emailOwners := make(map[string]uuid.UUID, len(accounts))
		for id, account := range accounts {
			if !sourceIDs[id] {
				return fmt.Errorf("target conflict user_id=%s field=extra_account", id)
			}
			emailOwners[account.normalizedEmail] = id
		}

		for _, record := range records {
			user, userExists := users[record.id]
			account, accountExists := accounts[record.id]
			if userExists {
				if user.deleted {
					return fmt.Errorf("target conflict user_id=%s field=deletion_state", record.id)
				}
				if accountExists && user.displayName != record.displayName {
					return fmt.Errorf("target conflict user_id=%s field=display_name", record.id)
				}
			}
			if accountExists {
				for _, field := range []struct {
					name string
					same bool
				}{
					{"email", account.email == record.email},
					{"email_normalized", account.normalizedEmail == record.normalizedEmail},
					{"password_hash", account.hash == record.hash},
					{"state", account.state == record.state},
					{"accounts.created_at", account.createdAt.Equal(record.createdAt)},
				} {
					if !field.same {
						return fmt.Errorf("target conflict user_id=%s field=%s", record.id, field.name)
					}
				}
			}
			if id, exists := emailOwners[record.normalizedEmail]; exists && id != record.id {
				return fmt.Errorf("target conflict user_id=%s field=email_normalized owner_differs", record.id)
			}
			if !userExists {
				newUsers++
			}
			if !accountExists {
				newAccounts++
			}
			updateName := userExists && user.displayName != record.displayName
			if updateName {
				updatedUsers++
			}
			if !userExists || !accountExists || updateName {
				_, err := fmt.Fprintf(
					output,
					"difference user_id=%s insert_user=%t insert_account=%t update_display_name=%t\n",
					record.id,
					!userExists,
					!accountExists,
					updateName,
				)
				if err != nil {
					return err
				}
			}
		}
		_, err = fmt.Fprintf(
			output,
			"dry_run=%t insert_users=%d insert_accounts=%d update_users=%d unchanged_accounts=%d\n",
			!apply,
			newUsers,
			newAccounts,
			updatedUsers,
			len(accounts),
		)
		if err != nil {
			return err
		}
		if !apply {
			return nil
		}

		for _, record := range records {
			user, exists := users[record.id]
			if !exists {
				if err := repository.insertImportedUser(ctx, record); err != nil {
					return databaseError("target write", err)
				}
			} else if user.displayName != record.displayName {
				if err := repository.updateImportedUserName(ctx, record); err != nil {
					return databaseError("target write", err)
				}
			}
			if _, exists := accounts[record.id]; !exists {
				if err := repository.insertImportedAccount(ctx, record); err != nil {
					return databaseError("target write", err)
				}
			}
		}
		count, err := repository.importAccountCount(ctx)
		if err != nil {
			return databaseError("target validation", err)
		}
		if count != int64(len(records)) {
			return errors.New("target validation: account count differs from source snapshot")
		}
		return nil
	})
	if err != nil {
		if operationErr != nil {
			return operationErr
		}
		return databaseError("target transaction (commit outcome may be uncertain)", err)
	}
	if apply {
		_, err = fmt.Fprintf(output, "committed_accounts=%d target_accounts=%d\n", newAccounts, len(records))
	}
	return err
}

func databaseError(stage string, err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return fmt.Errorf("%s failed (SQLSTATE %s); database details suppressed", stage, pgError.Code)
	}
	return fmt.Errorf("%s failed; connection and database details suppressed", stage)
}

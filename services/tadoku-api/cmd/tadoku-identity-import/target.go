package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type targetUser struct {
	id          uuid.UUID
	displayName string
	createdAt   time.Time
	deleted     bool
}

func importTarget(
	ctx context.Context,
	pool *pgxpool.Pool,
	records []sourceIdentity,
	apply bool,
	output io.Writer,
) error {
	ctx = tenant.WithKey(ctx, tenant.Production())
	newUsers, newAccounts := 0, 0
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

		users, accounts, err := readTarget(ctx, db)
		if err != nil {
			return err
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
				if user.displayName != record.displayName {
					return fmt.Errorf("target conflict user_id=%s field=display_name", record.id)
				}
				if !user.createdAt.Equal(record.createdAt) {
					return fmt.Errorf("target conflict user_id=%s field=users.created_at", record.id)
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
			if !userExists || !accountExists {
				_, err := fmt.Fprintf(output, "difference user_id=%s insert_user=%t insert_account=%t\n",
					record.id, !userExists, !accountExists)
				if err != nil {
					return err
				}
			}
		}
		if _, err := fmt.Fprintf(output, "dry_run=%t insert_users=%d insert_accounts=%d unchanged_accounts=%d\n",
			!apply, newUsers, newAccounts, len(accounts)); err != nil {
			return err
		}
		if !apply {
			return nil
		}

		for _, record := range records {
			if _, exists := users[record.id]; !exists {
				_, err := db.Exec(ctx, `insert into users (tenant, id, display_name, created_at, updated_at)
					values (current_setting('tadoku.tenant'), $1, $2,
					$3::timestamptz at time zone 'UTC', $4::timestamptz at time zone 'UTC')`,
					record.id, record.displayName, record.createdAt, record.updatedAt)
				if err != nil {
					return databaseError("target write", err)
				}
			}
			if _, exists := accounts[record.id]; !exists {
				_, err := db.Exec(ctx, `insert into accounts
					(tenant, user_id, email, email_normalized, password_hash, state, created_at, updated_at)
					values (current_setting('tadoku.tenant'), $1, $2, $3, $4, $5, $6, $7)`,
					record.id, record.email, record.normalizedEmail, record.hash, record.state, record.createdAt, record.updatedAt)
				if err != nil {
					return databaseError("target write", err)
				}
			}
		}
		var count int
		err = db.QueryRow(ctx, `select count(*) from accounts
			where tenant = current_setting('tadoku.tenant')`).Scan(&count)
		if err != nil {
			return databaseError("target validation", err)
		}
		if count != len(records) {
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

func readTarget(
	ctx context.Context,
	db postgres.DBTX,
) (map[uuid.UUID]targetUser, map[uuid.UUID]sourceIdentity, error) {
	rows, err := db.Query(ctx, `select id, display_name, created_at at time zone 'UTC',
		deleted_at is not null or deletion_locked_at is not null
		from users
		where tenant = current_setting('tadoku.tenant')`)
	if err != nil {
		return nil, nil, databaseError("target read", err)
	}
	defer rows.Close()
	users := make(map[uuid.UUID]targetUser)
	for rows.Next() {
		var user targetUser
		if err := rows.Scan(&user.id, &user.displayName, &user.createdAt, &user.deleted); err != nil {
			return nil, nil, databaseError("target read", err)
		}
		users[user.id] = user
	}
	if err := rows.Err(); err != nil {
		return nil, nil, databaseError("target read", err)
	}

	rows, err = db.Query(ctx, `select user_id, email, email_normalized, password_hash, state, created_at
		from accounts
		where tenant = current_setting('tadoku.tenant')`)
	if err != nil {
		return nil, nil, databaseError("target read", err)
	}
	defer rows.Close()
	accounts := make(map[uuid.UUID]sourceIdentity)
	for rows.Next() {
		var account sourceIdentity
		err := rows.Scan(
			&account.id, &account.email, &account.normalizedEmail, &account.hash, &account.state, &account.createdAt,
		)
		if err != nil {
			return nil, nil, databaseError("target read", err)
		}
		accounts[account.id] = account
	}
	if err := rows.Err(); err != nil {
		return nil, nil, databaseError("target read", err)
	}
	return users, accounts, nil
}

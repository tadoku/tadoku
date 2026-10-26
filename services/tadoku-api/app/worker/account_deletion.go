package worker

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/app/accountdeletion"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type AccountDeletion struct {
	DB            *pgxpool.Pool
	Profile       *profile.Service
	Logs          *logs.Service
	Contests      *contests.Service
	Relationships *permissions.Relationships
	Identity      *identity.Deletion
	Audit         *audit.Service
	ContestEvents *contestevents.Service
}

func (a *Application) DeleteAccount(ctx context.Context, job jobs.DeleteAccountV1) error {
	return nil
}

func FileAccountDeletion(
	ctx context.Context,
	db *pgxpool.Pool,
	filing *accountdeletion.Filing,
	request accountdeletion.Request,
) (int64, error) {
	var jobID int64
	err := postgres.RunInTransaction(ctx, db, func(ctx context.Context) error {
		var err error
		jobID, err = filing.File(ctx, request)
		return err
	})
	if err != nil {
		return 0, err
	}
	return jobID, nil
}

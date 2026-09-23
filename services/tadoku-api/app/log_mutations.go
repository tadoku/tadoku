package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type LogCreateParameters struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32
	LanguageCode    string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string
}

type LogUpdateParameters struct {
	ID              uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string
}

func (a *Application) CreateLog(ctx context.Context, p LogCreateParameters) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	caller := identity.FromContext(ctx)

	now := timex.Now()
	userID, err := a.profile.SynchronizeUser(ctx, caller, now)
	if err != nil {
		return nil, err
	}

	var id uuid.UUID
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, userID); err != nil {
			return err
		}

		input, err := a.scoring.NormalizeLogInput(logscore.Input{
			UnitID:          p.UnitID,
			UnitKey:         p.UnitKey,
			ActivityID:      p.ActivityID,
			LanguageCode:    p.LanguageCode,
			Amount:          p.Amount,
			DurationSeconds: p.DurationSeconds,
			Tags:            p.Tags,
		})
		if err != nil {
			return err
		}

		var eligible []logscore.Target
		if len(p.RegistrationIDs) > 0 {
			eligible, err = a.contests.SelectRegistrationsForScoring(ctx, userID, p.RegistrationIDs, input.LanguageCode, input.ActivityID)
			if err != nil {
				return err
			}
		}

		scored, err := a.scoring.ScoreLog(ctx, scoring.LogCreate, input, eligible)
		if err != nil {
			return err
		}
		id, err = a.logs.Create(ctx, userID, now, p.Description, scored)
		return err
	})
	if err != nil {
		return nil, err
	}
	return a.logs.FindLog(ctx, id, false)
}

func (a *Application) UpdateLog(ctx context.Context, p LogUpdateParameters) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}
	existing, err := a.logs.FindLog(ctx, p.ID, false)
	if err != nil {
		return nil, err
	}
	if callerID != existing.UserID {
		admin, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return nil, err
		}
		if !admin {
			return nil, errx.NewForbiddenError("forbidden")
		}
	}

	now := timex.Now()
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, existing.UserID); err != nil {
			return err
		}

		input := logscore.Input{
			UnitID:          p.UnitID,
			UnitKey:         p.UnitKey,
			ActivityID:      existing.Activity.ID,
			LanguageCode:    existing.LanguageCode,
			Amount:          p.Amount,
			DurationSeconds: p.DurationSeconds,
			Tags:            p.Tags,
		}
		eligible := a.logs.RegistrationsForRescoring(existing, now)
		scored, err := a.scoring.ScoreLog(ctx, scoring.LogUpdate, input, eligible)
		if err != nil {
			return err
		}
		return a.logs.Update(ctx, p.ID, existing.UserID, now, p.Description, scored)
	})
	if err != nil {
		return nil, err
	}
	return a.logs.FindLog(ctx, p.ID, false)
}

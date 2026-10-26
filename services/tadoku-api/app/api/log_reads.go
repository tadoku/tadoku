package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Log = logs.Log
type LogList = logs.LogList
type LogListParameters = logs.ListParameters
type ExportedLog = logs.ExportedLog

func (a *Application) FindLog(ctx context.Context, id uuid.UUID) (*Log, error) {
	actorID, hasActor := identity.ActorID(ctx)
	isAdmin := hasActor && a.roles.IsAdminOrFalse(ctx)
	var viewer logs.Viewer = logs.GuestViewer{}
	switch {
	case isAdmin:
		viewer = logs.AdminViewer{}
	case hasActor:
		viewer = logs.UserViewer{UserID: actorID}
	}

	return a.logs.FindLogForViewer(ctx, id, logs.FindForViewerParameters{
		Viewer:         viewer,
		IncludeDeleted: isAdmin,
	})
}

func (a *Application) ListUserLogs(
	ctx context.Context,
	parameters LogListParameters,
	createdOn *time.Time,
) (*LogList, error) {
	if parameters.IncludeDeleted && !a.roles.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("include_deleted requires administrator access")
	}
	return a.logs.ListUserLogs(ctx, parameters, createdOn)
}

func (a *Application) ListContestLogs(ctx context.Context, parameters LogListParameters) (*LogList, error) {
	if parameters.IncludeDeleted && !a.roles.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("include_deleted requires administrator access")
	}

	if err := a.contests.RequireExistingContest(ctx, parameters.ContestID); err != nil {
		return nil, err
	}
	return a.logs.ListContestLogs(ctx, parameters)
}

func (a *Application) ExportOwnLogs(ctx context.Context) ([]ExportedLog, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	actorID, err := identity.RequireActorID(ctx)
	if err != nil {
		return nil, err
	}

	return a.logs.ExportUserLogs(ctx, actorID)
}

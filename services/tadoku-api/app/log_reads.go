package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Log = logs.Log
type LogList = logs.LogList
type LogListParameters = logs.ListParameters

func (a *Application) FindLog(ctx context.Context, id uuid.UUID) (*Log, error) {
	actorID, hasActor := identity.ActorID(ctx)
	isAdmin := hasActor && a.permissions.IsAdminOrFalse(ctx)
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

func (a *Application) ListUserLogs(ctx context.Context, parameters LogListParameters) (*LogList, error) {
	if parameters.IncludeDeleted && !a.permissions.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("include_deleted requires administrator access")
	}
	return a.logs.ListUserLogs(ctx, parameters)
}

func (a *Application) ListContestLogs(ctx context.Context, parameters LogListParameters) (*LogList, error) {
	if parameters.IncludeDeleted && !a.permissions.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("include_deleted requires administrator access")
	}

	if err := a.contests.RequireExistingContest(ctx, parameters.ContestID); err != nil {
		return nil, err
	}
	return a.logs.ListContestLogs(ctx, parameters)
}

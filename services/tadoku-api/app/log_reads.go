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
	var callerID *uuid.UUID
	if caller := identity.FromContext(ctx); caller != nil && caller.Subject != "guest" {
		userID, err := caller.UUID()
		if err != nil {
			return nil, errx.NewUnauthorizedError("unauthorized")
		}
		callerID = &userID
	}

	isAdmin := callerID != nil && a.permissions.IsAdminOrFalse(ctx)
	log, err := a.logs.FindLog(ctx, id, isAdmin)
	if err != nil {
		return nil, err
	}

	isOwner := callerID != nil && log.UserID == *callerID
	if !isAdmin && !isOwner {
		log.Registrations = nil
	}
	return log, nil
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

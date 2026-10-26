package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
)

type ContestEventPage = contestevents.Page

func (a *Application) ListContestEvents(
	ctx context.Context,
	contestID uuid.UUID,
	cursor string,
	limit int,
) (*ContestEventPage, error) {
	if err := a.contests.RequireExistingContest(ctx, contestID); err != nil {
		return nil, err
	}
	return a.contestEvents.List(ctx, contestID, cursor, limit)
}

func (a *Application) FindContestLog(ctx context.Context, contestID, logID uuid.UUID) (*Log, error) {
	if err := a.contests.RequireExistingContest(ctx, contestID); err != nil {
		return nil, err
	}
	return a.logs.FindContestLog(ctx, contestID, logID)
}

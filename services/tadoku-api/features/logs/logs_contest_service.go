package logs

import (
	"context"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
)

type contestLogView struct {
	ContestID uuid.UUID
	Log       Log
}

func (s *Service) FindContestLog(ctx context.Context, contestID, logID uuid.UUID) (*Log, error) {
	views, err := s.logs.ContestLogViews(ctx, logID, &contestID)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, ErrLogNotFound
	}
	log := views[0].Log
	if err := hydrateLogActivity(&log); err != nil {
		return nil, err
	}
	return &log, nil
}

func contestLogChanges(logID uuid.UUID, before, after []contestLogView) []contestevent.Change {
	previous := make(map[uuid.UUID]Log, len(before))
	for _, view := range before {
		previous[view.ContestID] = view.Log
	}

	changes := make([]contestevent.Change, 0)
	remaining := make(map[uuid.UUID]struct{}, len(after))
	for _, view := range after {
		remaining[view.ContestID] = struct{}{}
		old, exists := previous[view.ContestID]
		kind := contestevent.Submitted
		if exists {
			if sameContestLogFields(old, view.Log) {
				continue
			}
			kind = contestevent.Updated
		}
		changes = append(changes, contestevent.Change{ContestID: view.ContestID, LogID: logID, Type: kind})
	}

	for _, view := range before {
		if _, exists := remaining[view.ContestID]; !exists {
			changes = append(changes, contestevent.Change{
				ContestID: view.ContestID,
				LogID:     logID,
				Type:      contestevent.Withdrawn,
			})
		}
	}
	return changes
}

func sameContestLogFields(a, b Log) bool {
	return a.Amount == b.Amount &&
		a.Modifier == b.Modifier &&
		a.Score == b.Score &&
		a.UnitName == b.UnitName &&
		reflect.DeepEqual(a.Description, b.Description) &&
		reflect.DeepEqual(a.DurationSeconds, b.DurationSeconds) &&
		slices.Equal(a.Tags, b.Tags)
}

package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestwindow"
)

// RemoveUser requires the caller's transaction and a user locked for deletion. It deletes the user's logs except
// their history in finished contests, which it keeps frozen without descriptions or tags. Like an ordinary mutation,
// it diffs each log's contest views before and after these writes to find its contest events.
func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID, now time.Time) (RemovalResult, error) {
	before, err := s.logs.UserContestLogViews(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}

	earliestUnfinishedContestEnd := contestwindow.EarliestUnfinishedContestEnd(now)
	if err := s.logs.DeleteUserContestLogsOutsideFinishedContests(ctx, userID, earliestUnfinishedContestEnd); err != nil {
		return RemovalResult{}, err
	}
	kept, err := s.logs.FreezeUserContestHistory(ctx, userID, now)
	if err != nil {
		return RemovalResult{}, err
	}
	if err := s.logs.DeleteUserTags(ctx, userID); err != nil {
		return RemovalResult{}, err
	}
	deleted, err := s.logs.DeleteUserLogsOutsideContests(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}
	after, err := s.logs.UserContestLogViews(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}

	return RemovalResult{Events: userContestLogChanges(before, after), DeletedLogs: deleted, KeptLogs: kept}, nil
}

func userContestLogChanges(before, after []contestLogView) []contestevent.Change {
	remaining := make(map[uuid.UUID][]contestLogView)
	for _, view := range after {
		remaining[view.Log.ID] = append(remaining[view.Log.ID], view)
	}

	var changes []contestevent.Change
	for start := 0; start < len(before); {
		logID := before[start].Log.ID
		end := start
		for end < len(before) && before[end].Log.ID == logID {
			end++
		}
		changes = append(changes, contestLogChanges(logID, before[start:end], remaining[logID])...)
		start = end
	}
	return changes
}

package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
)

// RemoveUser requires the caller's transaction and a user locked for deletion. It deletes the user's logs except
// their history in finished contests, which it keeps frozen without descriptions or tags. Like an ordinary mutation,
// it diffs each log's contest views before and after these writes to find its contest events.
func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID, now time.Time) (RemovalResult, error) {
	before, err := s.logs.UserContestLogViews(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}
	years, err := s.logs.UserOfficialYears(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}

	if err := s.logs.DeleteUserContestLogsOutsideFinishedContests(ctx, userID, now); err != nil {
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

	changes := userContestLogChanges(before, after)

	result := RemovalResult{Events: changes, DeletedLogs: deleted, KeptLogs: kept}
	invalidated := make(map[uuid.UUID]bool)
	for _, change := range changes {
		if change.Type == contestevent.Withdrawn && !invalidated[change.ContestID] {
			invalidated[change.ContestID] = true
			result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: change.ContestID})
		}
	}
	for _, year := range years {
		result.Jobs = append(result.Jobs, jobs.InvalidateOfficialLeaderboardV1{Year: year})
	}
	return result, nil
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

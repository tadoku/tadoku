package logs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
)

// RemoveUser requires the caller's transaction and a user locked for deletion. It deletes the user's logs except
// their history in finished contests, which it keeps frozen without descriptions or tags. Its events are the ones the
// contest log diff of ordinary mutations gives for these writes: a withdrawal for each log leaving an unfinished
// contest and an update for each kept log whose description or tags it clears.
func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID, now time.Time) (RemovalResult, error) {
	changes, err := s.logs.UserContestLogRemovals(ctx, userID, now)
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

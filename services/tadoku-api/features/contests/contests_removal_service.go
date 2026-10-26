package contests

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
)

// RemoveUser requires the caller's transaction and a user locked for deletion. It deletes the user's registrations
// in unfinished contests and keeps the contests the user owns under "Deleted organizer".
func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID, now time.Time) (RemovalResult, error) {
	contestIDs, err := s.contests.DeleteUnfinishedRegistrations(ctx, userID, now)
	if err != nil {
		return RemovalResult{}, err
	}
	kept, err := s.contests.CountRegistrations(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}
	owned, err := s.contests.RenameContestsOwnedBy(ctx, userID)
	if err != nil {
		return RemovalResult{}, err
	}

	result := RemovalResult{DeletedRegistrations: len(contestIDs), KeptRegistrations: kept, OwnedContests: owned}
	for _, contestID := range contestIDs {
		result.Jobs = append(result.Jobs, jobs.InvalidateContestLeaderboardV1{ContestID: contestID})
	}
	return result, nil
}

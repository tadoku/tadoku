package domain

import (
	"github.com/google/uuid"
)

// buildLeaderboardEntries converts a LeaderboardPage (from the store)
// into LeaderboardEntry values with ranks, ties, and display names.
// StartRank from the page is used as the rank of the first entry.
// HasPrevTie/HasNextTie handle tie detection at page boundaries.
func buildLeaderboardEntries(page LeaderboardPage, displayNames map[uuid.UUID]string) []LeaderboardEntry {
	if len(page.Scores) == 0 {
		return []LeaderboardEntry{}
	}

	entries := make([]LeaderboardEntry, len(page.Scores))
	currentRank := page.StartRank
	for i, s := range page.Scores {
		if i > 0 && s.Score < page.Scores[i-1].Score {
			currentRank = page.StartRank + i
		}
		entries[i] = LeaderboardEntry{
			Rank:            currentRank,
			UserID:          s.UserID,
			UserDisplayName: displayNames[s.UserID],
			Score:           float32(s.Score),
		}
	}

	// Mark ties: entries that share a rank with at least one other entry
	for i := range entries {
		if i > 0 && entries[i].Rank == entries[i-1].Rank {
			entries[i].IsTie = true
			entries[i-1].IsTie = true
		}
	}

	// Handle boundary ties
	if page.HasPrevTie && len(entries) > 0 {
		entries[0].IsTie = true
	}
	if page.HasNextTie && len(entries) > 0 {
		entries[len(entries)-1].IsTie = true
	}

	return entries
}

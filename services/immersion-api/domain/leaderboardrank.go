package domain

import (
	"github.com/google/uuid"
)

// buildLeaderboardEntries converts a slice of LeaderboardScore (from the store)
// into LeaderboardEntry values with ranks, ties, and display names.
// pageOffset is the number of entries on prior pages (page * pageSize),
// used to compute correct global ranks.
func buildLeaderboardEntries(scores []LeaderboardScore, displayNames map[uuid.UUID]string, pageOffset int) []LeaderboardEntry {
	if len(scores) == 0 {
		return []LeaderboardEntry{}
	}

	entries := make([]LeaderboardEntry, len(scores))
	currentRank := pageOffset + 1
	for i, s := range scores {
		if i > 0 && s.Score < scores[i-1].Score {
			currentRank = pageOffset + i + 1
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

	return entries
}

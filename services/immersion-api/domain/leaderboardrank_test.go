package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuildLeaderboardEntries_Empty(t *testing.T) {
	result := buildLeaderboardEntries(nil, nil, 0)
	assert.Equal(t, []LeaderboardEntry{}, result)
}

func TestBuildLeaderboardEntries_SingleEntry(t *testing.T) {
	userID := uuid.New()
	scores := []LeaderboardScore{{UserID: userID, Score: 100}}
	names := map[uuid.UUID]string{userID: "Alice"}

	result := buildLeaderboardEntries(scores, names, 0)

	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, "Alice", result[0].UserDisplayName)
	assert.Equal(t, float32(100), result[0].Score)
	assert.False(t, result[0].IsTie)
}

func TestBuildLeaderboardEntries_DistinctScores(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	scores := []LeaderboardScore{
		{UserID: u1, Score: 300},
		{UserID: u2, Score: 200},
		{UserID: u3, Score: 100},
	}
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(scores, names, 0)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 2, result[1].Rank)
	assert.Equal(t, 3, result[2].Rank)
	assert.False(t, result[0].IsTie)
	assert.False(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_Ties(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	scores := []LeaderboardScore{
		{UserID: u1, Score: 200},
		{UserID: u2, Score: 200},
		{UserID: u3, Score: 100},
	}
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(scores, names, 0)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 1, result[1].Rank)
	assert.Equal(t, 3, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_AllTied(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	scores := []LeaderboardScore{
		{UserID: u1, Score: 100},
		{UserID: u2, Score: 100},
		{UserID: u3, Score: 100},
	}
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(scores, names, 0)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 1, result[1].Rank)
	assert.Equal(t, 1, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.True(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_PageOffset(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	scores := []LeaderboardScore{
		{UserID: u1, Score: 50},
		{UserID: u2, Score: 25},
	}
	names := map[uuid.UUID]string{u1: "A", u2: "B"}

	// Simulating page 1 with pageSize 10 (offset = 10)
	result := buildLeaderboardEntries(scores, names, 10)

	assert.Equal(t, 11, result[0].Rank)
	assert.Equal(t, 12, result[1].Rank)
}

func TestBuildLeaderboardEntries_PageOffsetWithTie(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	scores := []LeaderboardScore{
		{UserID: u1, Score: 50},
		{UserID: u2, Score: 50},
		{UserID: u3, Score: 25},
	}
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(scores, names, 5)

	assert.Equal(t, 6, result[0].Rank)
	assert.Equal(t, 6, result[1].Rank)
	assert.Equal(t, 8, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_MissingDisplayName(t *testing.T) {
	userID := uuid.New()
	scores := []LeaderboardScore{{UserID: userID, Score: 100}}
	names := map[uuid.UUID]string{} // no name for this user

	result := buildLeaderboardEntries(scores, names, 0)

	assert.Equal(t, "", result[0].UserDisplayName)
}

package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuildLeaderboardEntries_Empty(t *testing.T) {
	result := buildLeaderboardEntries(LeaderboardPage{}, nil)
	assert.Equal(t, []LeaderboardEntry{}, result)
}

func TestBuildLeaderboardEntries_SingleEntry(t *testing.T) {
	userID := uuid.New()
	names := map[uuid.UUID]string{userID: "Alice"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores:    []LeaderboardScore{{UserID: userID, Score: 100}},
		StartRank: 1,
	}, names)

	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, "Alice", result[0].UserDisplayName)
	assert.Equal(t, float32(100), result[0].Score)
	assert.False(t, result[0].IsTie)
}

func TestBuildLeaderboardEntries_DistinctScores(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 300},
			{UserID: u2, Score: 200},
			{UserID: u3, Score: 100},
		},
		StartRank: 1,
	}, names)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 2, result[1].Rank)
	assert.Equal(t, 3, result[2].Rank)
	assert.False(t, result[0].IsTie)
	assert.False(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_Ties(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 200},
			{UserID: u2, Score: 200},
			{UserID: u3, Score: 100},
		},
		StartRank: 1,
	}, names)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 1, result[1].Rank)
	assert.Equal(t, 3, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_AllTied(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 100},
			{UserID: u2, Score: 100},
			{UserID: u3, Score: 100},
		},
		StartRank: 1,
	}, names)

	assert.Equal(t, 1, result[0].Rank)
	assert.Equal(t, 1, result[1].Rank)
	assert.Equal(t, 1, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.True(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_StartRankOffset(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 50},
			{UserID: u2, Score: 25},
		},
		StartRank: 11,
	}, names)

	assert.Equal(t, 11, result[0].Rank)
	assert.Equal(t, 12, result[1].Rank)
}

func TestBuildLeaderboardEntries_StartRankWithTie(t *testing.T) {
	u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B", u3: "C"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 50},
			{UserID: u2, Score: 50},
			{UserID: u3, Score: 25},
		},
		StartRank: 6,
	}, names)

	assert.Equal(t, 6, result[0].Rank)
	assert.Equal(t, 6, result[1].Rank)
	assert.Equal(t, 8, result[2].Rank)
	assert.True(t, result[0].IsTie)
	assert.True(t, result[1].IsTie)
	assert.False(t, result[2].IsTie)
}

func TestBuildLeaderboardEntries_MissingDisplayName(t *testing.T) {
	userID := uuid.New()
	names := map[uuid.UUID]string{} // no name for this user

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores:    []LeaderboardScore{{UserID: userID, Score: 100}},
		StartRank: 1,
	}, names)

	assert.Equal(t, "", result[0].UserDisplayName)
}

func TestBuildLeaderboardEntries_BoundaryTies(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 100},
			{UserID: u2, Score: 50},
		},
		StartRank:  5,
		HasPrevTie: true,
		HasNextTie: true,
	}, names)

	assert.True(t, result[0].IsTie, "first entry should be tied with previous page")
	assert.True(t, result[1].IsTie, "last entry should be tied with next page")
}

func TestBuildLeaderboardEntries_PrevTieOnly(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	names := map[uuid.UUID]string{u1: "A", u2: "B"}

	result := buildLeaderboardEntries(LeaderboardPage{
		Scores: []LeaderboardScore{
			{UserID: u1, Score: 100},
			{UserID: u2, Score: 50},
		},
		StartRank:  5,
		HasPrevTie: true,
		HasNextTie: false,
	}, names)

	assert.True(t, result[0].IsTie, "first entry should be tied with previous page")
	assert.False(t, result[1].IsTie, "last entry should not be tied")
}

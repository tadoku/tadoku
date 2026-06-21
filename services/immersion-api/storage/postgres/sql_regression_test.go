package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratedLogReadQueriesUseEffectiveScoreAndTrackingFields(t *testing.T) {
	baseEffectiveScore := "coalesce(logs.computed_score, logs.score)"
	contestEffectiveScore := "coalesce(contest_logs.computed_score, contest_logs.score)"

	baseLogQueries := []struct {
		name     string
		query    string
		expected string
	}{
		{"ListLogsForUser", listLogsForUser, baseEffectiveScore},
		{"FindLogByID", findLogByID, baseEffectiveScore},
		{"FetchScoresForProfile", fetchScoresForProfile, baseEffectiveScore},
		{"YearlyActivitySplitForUser", yearlyActivitySplitForUser, baseEffectiveScore},
		{"YearlyActivityForUser", yearlyActivityForUser, "coalesce(computed_score, score)"},
	}
	for _, tt := range baseLogQueries {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.query, tt.expected)
		})
	}

	contestLogQueries := map[string]string{
		"ListLogsForContest":                     listLogsForContest,
		"FindAttachedContestRegistrationsForLog": findAttachedContestRegistrationsForLog,
	}
	for name, query := range contestLogQueries {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, query, contestEffectiveScore)
		})
	}

	trackingFieldQueries := map[string]string{
		"ListLogsForUser":    listLogsForUser,
		"ListLogsForContest": listLogsForContest,
		"FindLogByID":        findLogByID,
	}
	for name, query := range trackingFieldQueries {
		t.Run(name+"TrackingFields", func(t *testing.T) {
			assert.Contains(t, query, "logs.unit_id")
			assert.Contains(t, query, "duration_seconds")
		})
	}
}

func TestGeneratedLeaderboardQueriesUseEffectiveScore(t *testing.T) {
	baseEffectiveScore := "coalesce(computed_score, score)"
	contestEffectiveScore := "coalesce(contest_logs.computed_score, contest_logs.score)"

	for name, query := range map[string]string{
		"LeaderboardForContest":       leaderboardForContest,
		"ContestLeaderboardAllScores": contestLeaderboardAllScores,
		"UserContestScore":            userContestScore,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, query, contestEffectiveScore)
			assert.NotContains(t, query, "sum(contest_logs.score)")
		})
	}

	for name, query := range map[string]string{
		"YearlyLeaderboard":          yearlyLeaderboard,
		"GlobalLeaderboard":          globalLeaderboard,
		"YearlyLeaderboardAllScores": yearlyLeaderboardAllScores,
		"GlobalLeaderboardAllScores": globalLeaderboardAllScores,
		"UserYearlyScore":            userYearlyScore,
		"UserGlobalScore":            userGlobalScore,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, query, baseEffectiveScore)
			assert.NotContains(t, query, "sum(score)")
			assert.NotContains(t, query, "having sum(score)")
		})
	}

	assert.Contains(t, yearlyLeaderboardAllScores, "having sum(coalesce(computed_score, score)) > 0")
	assert.Contains(t, globalLeaderboardAllScores, "having sum(coalesce(computed_score, score)) > 0")
}

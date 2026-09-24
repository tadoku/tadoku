// Package leaderboardoutbox defines the leaderboard outbox event types shared by
// the features that write events and the leaderboard feature that consumes them.
package leaderboardoutbox

type EventType string

const (
	RefreshContestScore   EventType = "refresh_contest_score"
	RefreshOfficialScores EventType = "refresh_official_scores"
)

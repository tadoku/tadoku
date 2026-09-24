package leaderboardoutbox

type EventType string

const (
	RefreshContestScore   EventType = "refresh_contest_score"
	RefreshOfficialScores EventType = "refresh_official_scores"
)

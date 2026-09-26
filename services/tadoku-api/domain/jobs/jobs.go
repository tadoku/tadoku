package jobs

import (
	"errors"

	"github.com/google/uuid"
)

type Type string

const (
	LeaderboardInvalidateContestV1  Type = "leaderboard.invalidate_contest.v1"
	LeaderboardInvalidateOfficialV1 Type = "leaderboard.invalidate_official.v1"
)

type Job interface {
	Type() Type
	Validate() error
	job()
}

type InvalidateContestLeaderboardV1 struct {
	ContestID uuid.UUID `json:"contest_id"`
}

func (InvalidateContestLeaderboardV1) Type() Type { return LeaderboardInvalidateContestV1 }
func (InvalidateContestLeaderboardV1) job()       {}
func (j InvalidateContestLeaderboardV1) Validate() error {
	if j.ContestID == uuid.Nil {
		return errors.New("contest_id must be a nonzero UUID")
	}
	return nil
}

type InvalidateOfficialLeaderboardV1 struct {
	Year int16 `json:"year"`
}

func (InvalidateOfficialLeaderboardV1) Type() Type { return LeaderboardInvalidateOfficialV1 }
func (InvalidateOfficialLeaderboardV1) job()       {}
func (j InvalidateOfficialLeaderboardV1) Validate() error {
	if j.Year < 1 || j.Year > 9999 {
		return errors.New("year must be between 1 and 9999")
	}
	return nil
}

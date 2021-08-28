package app

import (
	"github.com/tadoku/tadoku/services/tadoku-contest-api/interfaces/services"
)

// Services is a collection of all services
type Services struct {
	Health     services.HealthService
	Session    services.SessionService
	Contest    services.ContestService
	Ranking    services.RankingService
	ContestLog services.ContestLogService
}

// NewServices initializes all interactors
func NewServices(i *Interactors, sessionCookieName string) *Services {
	return &Services{
		Health:     services.NewHealthService(),
		Session:    services.NewSessionService(i.Session, sessionCookieName),
		Contest:    services.NewContestService(i.Contest),
		Ranking:    services.NewRankingService(i.Ranking),
		ContestLog: services.NewContestLogService(i.Ranking),
	}
}

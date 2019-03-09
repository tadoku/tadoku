package app

import (
	"github.com/tadoku/api/interfaces/services"
)

// Services is a collection of all services
type Services struct {
	Health  services.HealthService
	Session services.SessionService
	Contest services.ContestService
	Ranking services.RankingService
}

// NewServices initializes all interactors
func NewServices(i *Interactors) *Services {
	return &Services{
		Health:  services.NewHealthService(),
		Session: services.NewSessionService(i.Session),
		Contest: services.NewContestService(i.Contest),
		Ranking: services.NewRankingService(),
	}
}

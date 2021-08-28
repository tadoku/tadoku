package app

import (
	"github.com/tadoku/tadoku/services/reading-contest-api/interfaces/services"
)

// Services is a collection of all services
type Services struct {
	Health     services.HealthService
	Contest    services.ContestService
	Ranking    services.RankingService
	ContestLog services.ContestLogService
}

// NewServices initializes all interactors
func NewServices(i *Interactors) *Services {
	return &Services{
		Health:     services.NewHealthService(),
		Contest:    services.NewContestService(i.Contest),
		Ranking:    services.NewRankingService(i.Ranking),
		ContestLog: services.NewContestLogService(i.Ranking),
	}
}

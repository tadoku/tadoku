package app

import (
	"github.com/tadoku/tadoku/services/reading-contest-api/infra"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"
)

// Interactors is a collection of all repositories
type Interactors struct {
	Contest usecases.ContestInteractor
	Ranking usecases.RankingInteractor
}

// NewInteractors initializes all repositories
func NewInteractors(
	r *Repositories,
) *Interactors {
	return &Interactors{
		Contest: usecases.NewContestInteractor(r.Contest, infra.NewValidator()),
		Ranking: usecases.NewRankingInteractor(r.Ranking, r.Contest, r.ContestLog, infra.NewValidator()),
	}
}

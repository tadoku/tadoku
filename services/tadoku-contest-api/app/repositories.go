package app

import (
	"github.com/tadoku/api/interfaces/rdb"
	r "github.com/tadoku/api/interfaces/repositories"
	"github.com/tadoku/api/usecases"
)

// Repositories is a collection of all repositories
type Repositories struct {
	User       usecases.UserRepository
	Contest    usecases.ContestRepository
	ContestLog usecases.ContestLogRepository
	Ranking    usecases.RankingRepository
}

// NewRepositories initializes all repositories
func NewRepositories(sh rdb.SQLHandler) *Repositories {
	return &Repositories{
		User:       r.NewUserRepository(sh),
		Contest:    r.NewContestRepository(sh),
		ContestLog: r.NewContestLogRepository(sh),
		Ranking:    r.NewRankingRepository(sh),
	}
}

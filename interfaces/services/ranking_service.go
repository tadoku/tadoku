package services

import (
	"net/http"
)

// RankingService is responsible for managing rankings
type RankingService interface {
	Create(ctx Context) error
}

// NewRankingService initializer
func NewRankingService() RankingService {
	return &rankingService{}
}

type rankingService struct {
}

func (s *rankingService) Create(ctx Context) error {
	// TODO: implement this
	return ctx.NoContent(http.StatusCreated)
}

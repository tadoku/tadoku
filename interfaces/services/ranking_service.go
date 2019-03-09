package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
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

type CreateRankingPayload struct {
	ContestID uint64               `json:"contest_id"`
	Languages domain.LanguageCodes `json:"languages"`
}

func (s *rankingService) Create(ctx Context) error {
	// TODO: implement this
	return ctx.NoContent(http.StatusCreated)
}

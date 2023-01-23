package query

import (
	"context"

	"github.com/tadoku/tadoku/services/common/domain"
)

type FetchContestConfigurationOptionsResponse struct {
	Languages              []Language
	Activities             []Activity
	CanCreateOfficialRound bool
}

func (s *ServiceImpl) FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error) {
	res, err := s.r.FetchContestConfigurationOptions(ctx)
	if err != nil {
		return nil, err
	}

	res.CanCreateOfficialRound = domain.IsRole(ctx, domain.RoleAdmin)

	return res, nil
}

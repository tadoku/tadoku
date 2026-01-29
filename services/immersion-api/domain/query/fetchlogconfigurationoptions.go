package query

import (
	"context"

	"github.com/tadoku/tadoku/services/common/domain"
)

type FetchLogConfigurationOptionsResponse struct {
	Languages  []Language
	Activities []Activity
	Units      []Unit
}

func (s *ServiceImpl) FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	return s.r.FetchLogConfigurationOptions(ctx)
}

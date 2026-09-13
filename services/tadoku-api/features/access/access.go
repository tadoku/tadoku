// Package access supplies authorization facts from the existing Keto provider.
package access

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/common/client/keto"
)

type Service struct{ roles *roles.KetoService }

func NewService(readURL string) *Service {
	return &Service{roles: roles.NewKetoService(keto.NewReadClient(readURL), "app", "tadoku")}
}

func (s *Service) Banned(ctx context.Context, subject string) (bool, error) {
	claims, err := s.roles.ClaimsForSubject(ctx, subject)
	return claims.Banned, err
}

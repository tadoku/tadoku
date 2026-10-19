package tokens

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Service struct{}

var (
	_ Verifier = (*Service)(nil)
	_ Manager  = (*Service)(nil)
)

func NewService() *Service { return &Service{} }
func (*Service) Verify(context.Context, string, tenant.Deployment) (Principal, error) {
	return Principal{}, ErrNotImplemented
}
func (*Service) Create(context.Context, Create) (Issued, error) { return Issued{}, ErrNotImplemented }
func (*Service) List(context.Context) ([]Token, error)          { return nil, ErrNotImplemented }
func (*Service) Revoke(context.Context, uuid.UUID) error        { return ErrNotImplemented }

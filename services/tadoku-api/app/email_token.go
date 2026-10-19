package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (*Application) AuthenticateEmailReader(context.Context, string) (tenant.Key, error) {
	return tenant.Key{}, errx.NewNotImplementedError()
}

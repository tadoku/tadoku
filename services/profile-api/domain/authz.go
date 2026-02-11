package domain

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
)

func requireAdmin(ctx context.Context) error { return roles.RequireAdmin(ctx) }

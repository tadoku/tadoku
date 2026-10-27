package worker

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (r *runner) expireCredentials(ctx context.Context) {
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	keys, err := r.maintenanceTenants(bounded)
	if err != nil {
		if bounded.Err() == nil {
			r.logger.ErrorContext(ctx, "list credential expiry tenants", "error", err)
		}
		return
	}
	for _, key := range keys {
		tenantCtx := tenant.WithKey(bounded, key)
		if err := r.identity.RemoveExpired(tenantCtx); err != nil {
			if bounded.Err() != nil {
				return
			}
			r.logger.ErrorContext(tenantCtx, "remove expired credentials", "error", err)
		}
	}
}

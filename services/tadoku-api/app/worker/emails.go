package worker

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (r *runner) maintainEmails(ctx context.Context, retention bool) {
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	keys, err := r.maintenanceTenants(bounded)
	if err != nil {
		r.logger.ErrorContext(ctx, "list email maintenance tenants", "error", err)
		return
	}
	for _, key := range keys {
		tenantCtx := tenant.WithKey(bounded, key)
		if err := r.email.ReconcileFailures(tenantCtx); err != nil {
			r.logger.ErrorContext(tenantCtx, "reconcile failed emails", "error", err)
		}
		if retention {
			if err := r.email.Retain(tenantCtx); err != nil {
				r.logger.ErrorContext(tenantCtx, "retain emails", "error", err)
			}
		}
	}
}

func (r *runner) maintenanceTenants(ctx context.Context) ([]tenant.Key, error) {
	scoped, err := r.scope.Context(ctx)
	if err != nil {
		return nil, err
	}
	if key, ok := tenant.FromContext(scoped); ok {
		return []tenant.Key{key}, nil
	}
	return r.queue.Tenants(scoped)
}

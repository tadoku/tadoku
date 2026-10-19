package worker

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func (r *runner) maintainEmails(ctx context.Context, retention bool) {
	if r.email == nil {
		return
	}
	bounded, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	scoped, err := r.scope.Context(bounded)
	if err != nil {
		r.logger.ErrorContext(ctx, "scope email maintenance", "error", err)
		return
	}
	keys := []tenant.Key{}
	if key, ok := tenant.FromContext(scoped); ok {
		keys = append(keys, key)
	} else {
		keys, err = r.email.Tenants(scoped)
		if err != nil {
			r.logger.ErrorContext(ctx, "list email tenants", "error", err)
			return
		}
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

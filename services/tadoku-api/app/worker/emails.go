package worker

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
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
		if err := r.reconcileEmailFailures(tenantCtx); err != nil {
			r.logger.ErrorContext(tenantCtx, "reconcile failed emails", "error", err)
		}
		if retention {
			if err := r.email.Retain(tenantCtx); err != nil {
				r.logger.ErrorContext(tenantCtx, "retain emails", "error", err)
			}
		}
	}
}

func (r *runner) reconcileEmailFailures(ctx context.Context) error {
	failed, err := r.queue.Failures(ctx, jobs.EmailSendV1)
	if err != nil {
		return err
	}

	failures := make([]email.SendFailure, 0, len(failed))
	for _, job := range failed {
		var payload jobs.SendEmailV1
		if decode(job.Payload, &payload) != nil || payload.Validate() != nil {
			continue
		}
		failures = append(failures, email.SendFailure{EmailID: payload.EmailID, Error: job.Code})
	}
	return r.email.ReconcileFailures(ctx, failures)
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

package observability

import (
	"context"
	"log/slog"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type tenantHandler struct {
	slog.Handler
}

func NewTenantHandler(handler slog.Handler) slog.Handler {
	return tenantHandler{Handler: handler}
}

func (handler tenantHandler) Handle(ctx context.Context, record slog.Record) error {
	if key, ok := tenant.FromContext(ctx); ok {
		record = record.Clone()
		record.AddAttrs(slog.String("tenant", key.String()))
	}
	return handler.Handler.Handle(ctx, record)
}

func (handler tenantHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTenantHandler(handler.Handler.WithAttrs(attrs))
}

func (handler tenantHandler) WithGroup(name string) slog.Handler {
	return NewTenantHandler(handler.Handler.WithGroup(name))
}

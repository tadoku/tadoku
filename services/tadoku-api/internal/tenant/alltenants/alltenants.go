package alltenants

import "context"

type contextKey struct{}

func With(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, true)
}

func Enabled(ctx context.Context) bool {
	enabled, _ := ctx.Value(contextKey{}).(bool)
	return enabled
}

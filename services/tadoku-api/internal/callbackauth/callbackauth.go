package callbackauth

import "context"

type authenticatedKey struct{}

func WithAuthentication(ctx context.Context) context.Context {
	return context.WithValue(ctx, authenticatedKey{}, true)
}

func IsAuthenticated(ctx context.Context) bool {
	authenticated, _ := ctx.Value(authenticatedKey{}).(bool)
	return authenticated
}

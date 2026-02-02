package serviceauth

import (
	"context"
	"net/http"
)

// WithServiceAuth returns a request editor function that adds service authentication headers.
// The returned function is compatible with oapi-codegen generated clients' RequestEditorFn type.
//
// Usage with oapi-codegen generated clients:
//
//	client, _ := internalapi.NewClientWithResponses(
//	    "http://profile-api:8080",
//	    internalapi.WithRequestEditorFn(serviceauth.WithServiceAuth(generator, "profile-api")),
//	)
func WithServiceAuth(generator *TokenGenerator, targetService string) func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		token, err := generator.Generate(targetService)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

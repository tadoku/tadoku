package s2s

import (
	"fmt"
	"net/http"
)

type authTransport struct {
	base          http.RoundTripper
	client        *Client
	targetService string
}

func NewAuthTransport(client *Client, targetService string, base http.RoundTripper) http.RoundTripper {
	return &authTransport{
		base:          base,
		client:        client,
		targetService: targetService,
	}
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.client == nil {
		return nil, fmt.Errorf("s2s client is required")
	}
	if t.targetService == "" {
		return nil, fmt.Errorf("target service is required")
	}

	token, err := t.client.GetTokenContext(req.Context(), t.targetService)
	if err != nil {
		return nil, err
	}

	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+token)

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(clone)
}

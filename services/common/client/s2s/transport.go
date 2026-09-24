package s2s

import (
	"fmt"
	"net/http"
)

type AuthTransport struct {
	Base          http.RoundTripper
	Client        *Client
	TargetService string
}

func NewAuthTransport(client *Client, targetService string, base http.RoundTripper) http.RoundTripper {
	return &AuthTransport{
		Base:          base,
		Client:        client,
		TargetService: targetService,
	}
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		return nil, fmt.Errorf("s2s client is required")
	}
	if t.TargetService == "" {
		return nil, fmt.Errorf("target service is required")
	}

	token, err := t.Client.GetTokenContext(req.Context(), t.TargetService)
	if err != nil {
		return nil, err
	}

	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+token)

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(clone)
}

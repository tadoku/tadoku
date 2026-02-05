package s2s

import (
	"fmt"
	"net/http"
)

// AuthTransport attaches an S2S bearer token to every request.
type AuthTransport struct {
	Base          http.RoundTripper
	Client        *Client
	TargetService string
}

// NewAuthTransport returns a RoundTripper that injects S2S auth headers.
func NewAuthTransport(client *Client, targetService string, base http.RoundTripper) http.RoundTripper {
	return &AuthTransport{
		Base:          base,
		Client:        client,
		TargetService: targetService,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Client == nil {
		return nil, fmt.Errorf("s2s client is required")
	}
	if t.TargetService == "" {
		return nil, fmt.Errorf("target service is required")
	}

	token, err := t.Client.GetToken(t.TargetService)
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

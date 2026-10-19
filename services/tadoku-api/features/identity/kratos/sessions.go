package kratos

import (
	"context"
	"fmt"
	"net/http"

	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Sessions struct {
	public     *kratosapi.APIClient
	admin      *kratosapi.APIClient
	configured bool
}

func NewSessions(publicURL, adminURL string, client *http.Client) *Sessions {
	publicConfig, adminConfig := kratosapi.NewConfiguration(), kratosapi.NewConfiguration()
	publicConfig.Servers = kratosapi.ServerConfigurations{{URL: publicURL}}
	adminConfig.Servers = kratosapi.ServerConfigurations{{URL: adminURL}}
	publicConfig.HTTPClient, adminConfig.HTTPClient = client, client
	return &Sessions{
		public:     kratosapi.NewAPIClient(publicConfig),
		admin:      kratosapi.NewAPIClient(adminConfig),
		configured: publicURL != "",
	}
}

func (s *Sessions) Check(ctx context.Context, cookie string) (*kratosapi.Session, error) {
	if !s.configured {
		return nil, fmt.Errorf("Kratos public session URL is not configured")
	}
	session, response, err := s.public.FrontendApi.ToSession(ctx).Cookie(cookie).Execute()
	if err != nil {
		if response != nil && (response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden) {
			return nil, nil
		}
		return nil, fmt.Errorf("Kratos session check failed")
	}
	if !session.GetActive() {
		return nil, nil
	}
	return session, nil
}

func (s *Sessions) Revoke(ctx context.Context, sessionID string) error {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant required for Kratos session revocation")
	}
	if key != tenant.Production() {
		return nil
	}
	response, err := s.admin.IdentityApi.DisableSession(ctx, sessionID).Execute()
	if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
		return fmt.Errorf("Kratos session revocation failed")
	}
	return nil
}

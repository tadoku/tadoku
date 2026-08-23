package rest_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/proxyapi"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
)

type proxyAdminKeto struct {
	allowed bool
	err     error
	called  bool
}

func (m *proxyAdminKeto) CheckPermission(
	ctx context.Context,
	namespace string,
	object string,
	relation string,
	subject ketoclient.Subject,
) (bool, error) {
	m.called = true
	return m.allowed, m.err
}

func TestProxyAdminCheck(t *testing.T) {
	tests := []struct {
		name       string
		keto       *proxyAdminKeto
		body       string
		token      string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "allowed returns OK without a body",
			keto:       &proxyAdminKeto{allowed: true},
			body:       `{"subject":"6f75df3f-a162-4c1a-b206-5ec2c62b25a9"}`,
			token:      "test-oathkeeper-token",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "denied returns forbidden",
			keto:       &proxyAdminKeto{allowed: false},
			body:       `{"subject":"d1fc57e1-1654-4092-a044-15db9f894dd9"}`,
			token:      "test-oathkeeper-token",
			wantStatus: http.StatusForbidden,
			wantCalled: true,
		},
		{
			name:       "Keto unavailable returns service unavailable",
			keto:       &proxyAdminKeto{err: errors.New("keto unavailable")},
			body:       `{"subject":"723f68ec-2260-4c93-8a54-0443629a7285"}`,
			token:      "test-oathkeeper-token",
			wantStatus: http.StatusServiceUnavailable,
			wantCalled: true,
		},
		{
			name:       "missing subject returns bad request",
			keto:       &proxyAdminKeto{},
			body:       `{}`,
			token:      "test-oathkeeper-token",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed Kratos subject returns bad request",
			keto:       &proxyAdminKeto{},
			body:       `{"subject":"forged-subject"}`,
			token:      "test-oathkeeper-token",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing callback token returns unauthorized",
			keto:       &proxyAdminKeto{allowed: true},
			body:       `{"subject":"6f75df3f-a162-4c1a-b206-5ec2c62b25a9"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong callback token returns unauthorized",
			keto:       &proxyAdminKeto{allowed: true},
			body:       `{"subject":"6f75df3f-a162-4c1a-b206-5ec2c62b25a9"}`,
			token:      "wrong-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			server := rest.NewServer(nil, nil, nil, nil, nil, domain.NewProxyAdminCheck(tt.keto))
			proxy := e.Group("", rest.OathkeeperAuthorization("test-oathkeeper-token"))
			proxyapi.RegisterHandlers(proxy, server)

			req := httptest.NewRequest(http.MethodPost, "/internal/v1/proxy/admin-check", bytes.NewBufferString(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			}
			resp := httptest.NewRecorder()

			e.ServeHTTP(resp, req)

			assert.Equal(t, tt.wantStatus, resp.Code)
			assert.Equal(t, tt.wantCalled, tt.keto.called)
			if tt.wantStatus == http.StatusOK {
				require.Empty(t, resp.Body.String())
			}
		})
	}
}

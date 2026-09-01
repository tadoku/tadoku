package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

type featureAccessStoreStub struct {
	state domain.FeatureAccessState
	err   error
}

func (s *featureAccessStoreStub) GetNamedUserAccess(context.Context, domain.FeatureFlagKey, uuid.UUID) (domain.FeatureAccessState, error) {
	return s.state, s.err
}

func (s *featureAccessStoreStub) SetNamedUserAccess(_ context.Context, _ domain.FeatureFlagKey, _ uuid.UUID, enabled bool) (domain.FeatureAccessState, error) {
	result := s.state
	result.Enabled = enabled
	return result, s.err
}

type featureAccessAuditStub struct{}

func (*featureAccessAuditStub) CreateModerationAuditLog(context.Context, *domain.ModerationAuditLogCreateRequest) error {
	return nil
}

func TestFeatureAccessHandlersReturnTypedNoStoreResponse(t *testing.T) {
	state := domain.FeatureAccessState{
		Enabled: true, Changed: true, Environment: "production", Revision: strings.Repeat("a", 40),
	}
	service := domain.NewFeatureAccess(&featureAccessStoreStub{state: state}, &featureAccessAuditStub{})
	server := &Server{featureAccess: service}
	targetID := uuid.New()

	for _, test := range []struct {
		name string
		call func(echo.Context) error
	}{
		{name: "get", call: func(ctx echo.Context) error {
			return server.FeatureAccessGet(ctx, openapi.ManagedFeatureFlagKey("release-log-entry-v2"), targetID)
		}},
		{name: "grant", call: func(ctx echo.Context) error {
			return server.FeatureAccessGrant(ctx, openapi.ManagedFeatureFlagKey("release-log-entry-v2"), targetID)
		}},
		{name: "revoke", call: func(ctx echo.Context) error {
			return server.FeatureAccessRevoke(ctx, openapi.ManagedFeatureFlagKey("release-log-entry-v2"), targetID)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(authzctx.AdminSubject(uuid.NewString()))
			recorder := httptest.NewRecorder()

			require.NoError(t, test.call(e.NewContext(req, recorder)))
			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, "private, no-store", recorder.Header().Get(echo.HeaderCacheControl))
			var response openapi.FeatureAccessResponse
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			assert.Equal(t, state.Revision, response.Revision)
		})
	}
}

func TestFeatureAccessHandlerSanitizesUpstreamFailure(t *testing.T) {
	service := domain.NewFeatureAccess(&featureAccessStoreStub{
		err: fmt.Errorf("%w: private details", domain.ErrFeatureAccessUnavailable),
	}, &featureAccessAuditStub{})
	server := &Server{featureAccess: service}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(authzctx.AdminSubject(uuid.NewString()))
	recorder := httptest.NewRecorder()

	require.NoError(t, server.FeatureAccessGet(e.NewContext(req, recorder), openapi.ManagedFeatureFlagKey("release-log-entry-v2"), uuid.New()))
	assert.Equal(t, http.StatusBadGateway, recorder.Code)
	assert.Empty(t, recorder.Body.String())
}

package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/featureflags"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
)

type publicFeatureFlagEvaluatorStub struct {
	decisions featureflags.PublicDecisions
	user      *commondomain.UserIdentity
}

func (s *publicFeatureFlagEvaluatorStub) EvaluatePublic(_ context.Context, user *commondomain.UserIdentity) featureflags.PublicDecisions {
	s.user = user
	return s.decisions
}

func TestFeatureFlagDecisionsReturnsPrivateAllowlistedResponse(t *testing.T) {
	subject := uuid.NewString()
	evaluator := &publicFeatureFlagEvaluatorStub{
		decisions: featureflags.PublicDecisions{ReleaseLogEntryV2: true},
	}
	server := &Server{featureFlagDecisions: evaluator}
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/feature-flags", nil)
	request = request.WithContext(authzctx.UserSubject(subject))
	recorder := httptest.NewRecorder()

	err := server.FeatureFlagDecisions(e.NewContext(request, recorder))

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "private, no-store", recorder.Header().Get(echo.HeaderCacheControl))
	require.NotNil(t, evaluator.user)
	assert.Equal(t, subject, evaluator.user.Subject)

	var response struct {
		Decisions map[string]bool `json:"decisions"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, map[string]bool{"release-log-entry-v2": true}, response.Decisions)
}

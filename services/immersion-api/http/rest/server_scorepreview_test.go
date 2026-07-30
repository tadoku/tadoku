package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

type scorePreviewRepositoryStub struct {
	ruleSet *domain.ScoringRuleSet
}

func (s *scorePreviewRepositoryStub) FetchOngoingContestRegistrations(context.Context, *domain.RegistrationListOngoingRequest) (*domain.ContestRegistrations, error) {
	return &domain.ContestRegistrations{}, nil
}

func (s *scorePreviewRepositoryStub) FindUnitForTracking(_ context.Context, req *domain.UnitFindForTrackingRequest) (*domain.Unit, error) {
	return &domain.Unit{
		ID:            req.ID,
		Key:           domain.UnitKeyReadingPage,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (s *scorePreviewRepositoryStub) FindUnitForTrackingByKey(_ context.Context, req *domain.UnitFindForTrackingByKeyRequest) (*domain.Unit, error) {
	return &domain.Unit{
		ID:            uuid.New(),
		Key:           req.Key,
		LogActivityID: int(req.ActivityID),
		Modifier:      1,
		LanguageCode:  &req.LanguageCode,
	}, nil
}

func (s *scorePreviewRepositoryStub) FindActivePlatformScoringRuleSet(context.Context) (*domain.ScoringRuleSet, error) {
	return s.ruleSet, nil
}

func (s *scorePreviewRepositoryStub) FindContestScoringRuleSets(context.Context, uuid.UUID) (*domain.ScoringRuleSet, *domain.ScoringRuleSet, error) {
	return nil, nil, nil
}

func TestScorePreviewReturnsAppliedRules(t *testing.T) {
	ruleSetID := uuid.New()
	ruleID := uuid.New()
	service := domain.NewScorePreview(&scorePreviewRepositoryStub{
		ruleSet: &domain.ScoringRuleSet{
			ID: ruleSetID,
			Rules: []domain.ScoringRule{{
				ID:          ruleID,
				Priority:    1,
				ActivityID:  1,
				UnitKey:     domain.UnitKeyReadingPage,
				ScoreSource: domain.ScoreSourceAmount,
				Rate:        2,
			}},
		},
	}, commondomain.NewMockClock(time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)))
	server := &Server{scorePreview: service}
	echoServer := echo.New()
	request := httptest.NewRequest(
		http.MethodPost,
		"/logs/score-preview",
		strings.NewReader(`{"activity_id":1,"language_code":"jpn","unit_key":"reading_page","amount":10,"tags":[]}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	request = request.WithContext(authzctx.UserSubject(uuid.NewString()))
	recorder := httptest.NewRecorder()
	ctx := echoServer.NewContext(request, recorder)

	err := server.ScorePreview(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, recorder.Code)
	var response openapi.ScorePreview
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, float32(20), response.Platform.Score)
	assert.Equal(t, &ruleSetID, response.Platform.RuleSetId)
	require.Len(t, response.Platform.Rules, 1)
	assert.Equal(t, ruleID, response.Platform.Rules[0].RuleId)
	assert.Equal(t, float32(2), response.Platform.Rules[0].Rate)
	assert.Empty(t, response.Contests)
}

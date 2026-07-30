package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func TestScoringRuleSetCreatePlatformRejectsInvalidJSON(t *testing.T) {
	echoServer := echo.New()
	request := httptest.NewRequest(http.MethodPost, "/scoring/rule-sets", strings.NewReader("{"))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := echoServer.NewContext(request, recorder)
	server := &Server{}

	err := server.ScoringRuleSetCreatePlatform(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestScoringRuleSetToAPIIncludesVersionAndMatchers(t *testing.T) {
	ruleSetID := uuid.New()
	contestID := uuid.New()
	ruleID := uuid.New()
	createdAt := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)

	result := scoringRuleSetToAPI(domain.ScoringRuleSet{
		ID:        ruleSetID,
		Scope:     domain.ScoringRuleSetScopeContest,
		ContestID: &contestID,
		Version:   3,
		Status:    domain.ScoringRuleSetStatusDraft,
		Mode:      domain.ScoringRuleSetModeReplace,
		CreatedAt: createdAt,
		Rules: []domain.ScoringRule{{
			ID:           ruleID,
			Priority:     10,
			ActivityID:   1,
			UnitKey:      domain.UnitKeyReadingPage,
			LanguageCode: "jpn",
			Tag:          "book",
			ScoreSource:  domain.ScoreSourceAmount,
			Rate:         2,
		}},
	})

	assert.Equal(t, ruleSetID, result.Id)
	assert.Equal(t, &contestID, result.ContestId)
	assert.Equal(t, int32(3), result.Version)
	assert.Equal(t, openapi.ScoringRuleSetStatus(domain.ScoringRuleSetStatusDraft), result.Status)
	require.Len(t, result.Rules, 1)
	assert.Equal(t, &ruleID, result.Rules[0].Id)
	assert.Equal(t, optionalString(domain.UnitKeyReadingPage), result.Rules[0].UnitKey)
	assert.Equal(t, optionalString("jpn"), result.Rules[0].LanguageCode)
	assert.Equal(t, optionalString("book"), result.Rules[0].Tag)
}

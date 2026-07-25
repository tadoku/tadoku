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
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

type contestLogListRepositoryStub struct {
	response *domain.LogListForContestResponse
}

func (s *contestLogListRepositoryStub) ListLogsForContest(
	context.Context,
	*domain.LogListForContestRequest,
) (*domain.LogListForContestResponse, error) {
	return s.response, nil
}

type profileLogListRepositoryStub struct {
	response *domain.LogListForUserResponse
}

func (s *profileLogListRepositoryStub) ListLogsForUser(
	context.Context,
	*domain.LogListForUserRequest,
) (*domain.LogListForUserResponse, error) {
	return s.response, nil
}

func TestContestListLogsIncludesDuration(t *testing.T) {
	durationSeconds := int32(3600)
	service := domain.NewLogListForContest(&contestLogListRepositoryStub{
		response: &domain.LogListForContestResponse{
			Logs: []domain.Log{durationOnlyLog(durationSeconds)},
		},
	})
	server := &Server{logListForContest: service}

	response := executeLogListRequest(t, func(ctx echo.Context) error {
		return server.ContestListLogs(
			ctx,
			uuid.New(),
			openapi.ContestListLogsParams{},
		)
	})

	require.Len(t, response.Logs, 1)
	require.NotNil(t, response.Logs[0].DurationSeconds)
	assert.Equal(t, durationSeconds, *response.Logs[0].DurationSeconds)
}

func TestProfileListLogsIncludesDuration(t *testing.T) {
	durationSeconds := int32(3600)
	service := domain.NewLogListForUser(&profileLogListRepositoryStub{
		response: &domain.LogListForUserResponse{
			Logs: []domain.Log{durationOnlyLog(durationSeconds)},
		},
	})
	server := &Server{logListForUser: service}

	response := executeLogListRequest(t, func(ctx echo.Context) error {
		return server.ProfileListLogs(
			ctx,
			uuid.New(),
			openapi.ProfileListLogsParams{},
		)
	})

	require.Len(t, response.Logs, 1)
	require.NotNil(t, response.Logs[0].DurationSeconds)
	assert.Equal(t, durationSeconds, *response.Logs[0].DurationSeconds)
}

func durationOnlyLog(durationSeconds int32) domain.Log {
	return domain.Log{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		LanguageCode:    "jpn",
		LanguageName:    "Japanese",
		ActivityID:      2,
		Score:           30,
		DurationSeconds: &durationSeconds,
	}
}

func executeLogListRequest(
	t *testing.T,
	handler func(echo.Context) error,
) openapi.Logs {
	t.Helper()

	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	require.NoError(t, handler(e.NewContext(request, recorder)))
	require.Equal(t, http.StatusOK, recorder.Code)

	var response openapi.Logs
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

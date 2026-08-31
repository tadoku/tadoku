package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

func TestTokenResponseUsesReflectedJWTExpiry(t *testing.T) {
	now := time.Date(2026, 8, 31, 16, 0, 0, 0, time.UTC)
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, now.Add(15*time.Minute).Unix())))
	token := "header." + payload + ".signature"

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Id-Token", token)
	recorder := httptest.NewRecorder()

	newTokenHandler(commondomain.NewMockClock(now)).ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response TokenResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	assert.Equal(t, token, response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 900, response.ExpiresIn)
}

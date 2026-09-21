package main

import (
	"bytes"
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

func TestServiceAccountAuthorizer(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{
			name:   "exact principal",
			body:   `{"subject":"system:serviceaccount:tdk-dev-tadoku-api:tadoku-api","expected_subject":"system:serviceaccount:tdk-dev-tadoku-api:tadoku-api"}`,
			status: http.StatusOK,
		},
		{
			name:   "wrong namespace",
			body:   `{"subject":"system:serviceaccount:other:tadoku-api","expected_subject":"system:serviceaccount:tdk-dev-tadoku-api:tadoku-api"}`,
			status: http.StatusForbidden,
		},
		{
			name:   "wrong service account",
			body:   `{"subject":"system:serviceaccount:tdk-dev-tadoku-api:other","expected_subject":"system:serviceaccount:tdk-dev-tadoku-api:tadoku-api"}`,
			status: http.StatusForbidden,
		},
		{
			name:   "malformed principal",
			body:   `{"subject":"tadoku-api","expected_subject":"tadoku-api"}`,
			status: http.StatusForbidden,
		},
		{
			name:   "missing subject",
			body:   `{"expected_subject":"system:serviceaccount:tdk-dev-tadoku-api:tadoku-api"}`,
			status: http.StatusForbidden,
		},
		{
			name:   "invalid JSON",
			body:   `{`,
			status: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/authorize-service-account", bytes.NewBufferString(test.body))
			recorder := httptest.NewRecorder()

			newServiceAccountAuthorizer().ServeHTTP(recorder, request)

			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

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

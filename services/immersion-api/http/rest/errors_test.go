package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestHandleCommonErrorsReturnsAccountDeletionCode(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	handled, err := handleCommonErrors(
		e.NewContext(httptest.NewRequest(http.MethodPost, "/", nil), recorder),
		fmt.Errorf("mutation rejected: %w", domain.ErrAccountDeletionInProgress),
	)

	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, http.StatusConflict, recorder.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "account_deletion_in_progress", body["error"])
}

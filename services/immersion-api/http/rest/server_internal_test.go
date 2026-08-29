package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type accountDeletionLockerMock struct {
	req *domain.AccountDeletionLockRequest
	err error
}

func (m *accountDeletionLockerMock) Execute(_ context.Context, req *domain.AccountDeletionLockRequest) error {
	m.req = req
	return m.err
}

func TestInternalAccountDeletionLock(t *testing.T) {
	userID := uuid.New()
	requestID := uuid.New()
	locker := &accountDeletionLockerMock{}
	server := NewInternalServer(locker)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/account-deletion-locks", strings.NewReader(
		`{"user_id":"`+userID.String()+`","request_id":"`+requestID.String()+`"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	err := server.InternalAccountDeletionLock(e.NewContext(req, recorder))

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	require.NotNil(t, locker.req)
	assert.Equal(t, userID, locker.req.UserID)
	assert.Equal(t, requestID, locker.req.RequestID)
}

func TestInternalAccountDeletionLockMapsDomainErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
	}{
		"invalid":    {domain.ErrRequestInvalid, http.StatusBadRequest},
		"forbidden":  {domain.ErrForbidden, http.StatusForbidden},
		"unexpected": {errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			server := NewInternalServer(&accountDeletionLockerMock{err: tc.err})
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
				`{"user_id":"`+uuid.NewString()+`","request_id":"`+uuid.NewString()+`"}`,
			))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			require.NoError(t, server.InternalAccountDeletionLock(e.NewContext(req, recorder)))
			assert.Equal(t, tc.status, recorder.Code)
		})
	}
}

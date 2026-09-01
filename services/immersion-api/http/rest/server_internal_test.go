package rest

import (
	"context"
	"errors"
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
)

type accountDeletionLockerMock struct {
	req *domain.AccountDeletionLockRequest
	err error
}

type accountDeletionScrubberMock struct {
	req *domain.AccountDeletionScrubRequest
	err error
}

type accountDeletionEligibilityCheckerMock struct {
	req    *domain.AccountDeletionEligibilityRequest
	result *domain.AccountDeletionEligibilityResult
	err    error
}

func (m *accountDeletionEligibilityCheckerMock) Execute(_ context.Context, req *domain.AccountDeletionEligibilityRequest) (*domain.AccountDeletionEligibilityResult, error) {
	m.req = req
	if m.result == nil {
		m.result = &domain.AccountDeletionEligibilityResult{}
	}
	return m.result, m.err
}

func (m *accountDeletionScrubberMock) Execute(_ context.Context, req *domain.AccountDeletionScrubRequest) error {
	m.req = req
	return m.err
}

func (m *accountDeletionLockerMock) Execute(_ context.Context, req *domain.AccountDeletionLockRequest) error {
	m.req = req
	return m.err
}

func TestInternalAccountDeletionLock(t *testing.T) {
	userID := uuid.New()
	requestID := uuid.New()
	locker := &accountDeletionLockerMock{}
	server := NewInternalServer(locker, &accountDeletionScrubberMock{}, &accountDeletionEligibilityCheckerMock{})
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
		"invalid":         {domain.ErrRequestInvalid, http.StatusBadRequest},
		"forbidden":       {domain.ErrForbidden, http.StatusForbidden},
		"running contest": {domain.ErrRunningContestOwned, http.StatusConflict},
		"unexpected":      {errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			server := NewInternalServer(&accountDeletionLockerMock{err: tc.err}, &accountDeletionScrubberMock{}, &accountDeletionEligibilityCheckerMock{})
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

func TestInternalAccountDeletionScrub(t *testing.T) {
	userID := uuid.New()
	requestID := uuid.New()
	scrubber := &accountDeletionScrubberMock{}
	server := NewInternalServer(&accountDeletionLockerMock{}, scrubber, &accountDeletionEligibilityCheckerMock{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/account-deletion-scrubs", strings.NewReader(
		`{"user_id":"`+userID.String()+`","request_id":"`+requestID.String()+`"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	require.NoError(t, server.InternalAccountDeletionScrub(e.NewContext(req, recorder)))
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	require.NotNil(t, scrubber.req)
	assert.Equal(t, userID, scrubber.req.UserID)
	assert.Equal(t, requestID, scrubber.req.RequestID)
}

func TestInternalAccountDeletionScrubMapsConflicts(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		code string
	}{
		"not locked":      {domain.ErrAccountDeletionNotLocked, "account_deletion_not_locked"},
		"running contest": {domain.ErrRunningContestOwned, "running_contest_owned"},
	} {
		t.Run(name, func(t *testing.T) {
			server := NewInternalServer(&accountDeletionLockerMock{}, &accountDeletionScrubberMock{err: tc.err}, &accountDeletionEligibilityCheckerMock{})
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
				`{"user_id":"`+uuid.NewString()+`","request_id":"`+uuid.NewString()+`"}`,
			))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			require.NoError(t, server.InternalAccountDeletionScrub(e.NewContext(req, recorder)))
			assert.Equal(t, http.StatusConflict, recorder.Code)
			assert.JSONEq(t, `{"error":"`+tc.code+`"}`, recorder.Body.String())
		})
	}
}

func TestInternalAccountDeletionEligibility(t *testing.T) {
	userID := uuid.New()
	checker := &accountDeletionEligibilityCheckerMock{}
	server := NewInternalServer(&accountDeletionLockerMock{}, &accountDeletionScrubberMock{}, checker)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/account-deletion-eligibility", strings.NewReader(
		`{"user_id":"`+userID.String()+`"}`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	require.NoError(t, server.InternalAccountDeletionEligibility(e.NewContext(req, recorder)))
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	require.NotNil(t, checker.req)
	assert.Equal(t, userID, checker.req.UserID)
}

func TestInternalAccountDeletionEligibilityReturnsRunningOwnerConflict(t *testing.T) {
	availableAfter := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	checker := &accountDeletionEligibilityCheckerMock{result: &domain.AccountDeletionEligibilityResult{AvailableAfter: &availableAfter}}
	server := NewInternalServer(&accountDeletionLockerMock{}, &accountDeletionScrubberMock{}, checker)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"user_id":"`+uuid.NewString()+`"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()

	require.NoError(t, server.InternalAccountDeletionEligibility(e.NewContext(req, recorder)))
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.JSONEq(t, `{"error":"running_contest_owned","available_after":"2026-09-05T00:00:00Z"}`, recorder.Body.String())
}

func TestInternalAccountDeletionEligibilityMapsErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		err    error
		status int
	}{
		"invalid":    {domain.ErrRequestInvalid, http.StatusBadRequest},
		"forbidden":  {domain.ErrForbidden, http.StatusForbidden},
		"unexpected": {errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			server := NewInternalServer(
				&accountDeletionLockerMock{},
				&accountDeletionScrubberMock{},
				&accountDeletionEligibilityCheckerMock{err: tc.err},
			)
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"user_id":"`+uuid.NewString()+`"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()

			require.NoError(t, server.InternalAccountDeletionEligibility(e.NewContext(req, recorder)))
			assert.Equal(t, tc.status, recorder.Code)
		})
	}
}

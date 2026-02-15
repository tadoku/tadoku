package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChecker struct {
	name string
	err  error
}

func (c *mockChecker) Name() string                  { return c.name }
func (c *mockChecker) Check(_ context.Context) error { return c.err }

type slowChecker struct {
	name  string
	delay time.Duration
}

func (c *slowChecker) Name() string { return c.name }
func (c *slowChecker) Check(ctx context.Context) error {
	select {
	case <-time.After(c.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestLivezHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := LivezHandler(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestReadyzHandler_AllHealthy(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	checkers := []HealthChecker{
		&mockChecker{name: "postgres", err: nil},
		&mockChecker{name: "redis", err: nil},
	}

	err := ReadyzHandler(checkers)(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ReadyzResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "ready", response.Status)
	require.Len(t, response.Checks, 2)
	assert.Equal(t, "postgres", response.Checks[0].Name)
	assert.Equal(t, "up", response.Checks[0].Status)
	assert.Equal(t, "redis", response.Checks[1].Name)
	assert.Equal(t, "up", response.Checks[1].Status)
}

func TestReadyzHandler_OneUnhealthy(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	checkers := []HealthChecker{
		&mockChecker{name: "postgres", err: nil},
		&mockChecker{name: "redis", err: errors.New("connection refused")},
	}

	err := ReadyzHandler(checkers)(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response ReadyzResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "not_ready", response.Status)
	require.Len(t, response.Checks, 2)
	assert.Equal(t, "up", response.Checks[0].Status)
	assert.Equal(t, "down", response.Checks[1].Status)
	assert.Equal(t, "connection refused", response.Checks[1].Error)
}

func TestReadyzHandler_TimeoutEnforced(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	checkers := []HealthChecker{
		&slowChecker{name: "slow-db", delay: 10 * time.Second},
	}

	start := time.Now()
	err := ReadyzHandler(checkers)(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Less(t, elapsed, 5*time.Second)

	var response ReadyzResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "not_ready", response.Status)
	assert.Equal(t, "down", response.Checks[0].Status)
}

func TestReadyzHandler_NoCheckers(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := ReadyzHandler(nil)(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ReadyzResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "ready", response.Status)
}

func TestNewPostgresChecker_Name(t *testing.T) {
	checker := NewPostgresChecker("my-db", nil)
	assert.Equal(t, "my-db", checker.Name())
}

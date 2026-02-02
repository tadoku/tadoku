package rest

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Checks if service is responsive
// (GET /ping)
func (s *Server) Ping(ctx echo.Context) error {
	// If profile-api client is configured, call it for testing service auth
	if s.profileClient != nil {
		resp, err := s.profileClient.InternalPingWithResponse(ctx.Request().Context())
		if err != nil {
			return ctx.String(http.StatusOK, fmt.Sprintf("pong (profile-api error: %v)", err))
		}
		if resp.JSON200 != nil {
			return ctx.String(http.StatusOK, fmt.Sprintf("pong (profile-api: %s, caller: %s)", resp.JSON200.Status, resp.JSON200.Caller))
		}
		return ctx.String(http.StatusOK, fmt.Sprintf("pong (profile-api status: %d)", resp.StatusCode()))
	}
	return ctx.String(http.StatusOK, "pong")
}

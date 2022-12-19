package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Checks if service is responsive
// (GET /ping)
func (s *Server) Ping(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "pong")
}

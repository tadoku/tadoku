package rest

import (
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
)

// Fetches the summary for a contest
// (GET /contests/{id}/summary)
func (s *Server) ContestFetchSummary(ctx echo.Context, id types.UUID) error {
	return nil
}

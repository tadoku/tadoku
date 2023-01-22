package rest

import (
	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
)

// COMMANDS

// QUERIES

func (s *Server) ProfileFindByUserID(ctx echo.Context, userId types.UUID) error {
	return nil
}

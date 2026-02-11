package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/profile-api/http/rest/openapi/internalapi"
)

// InternalServer handles internal service-to-service endpoints
type InternalServer struct{}

func NewInternalServer() *InternalServer {
	return &InternalServer{}
}

// Ensure InternalServer implements the generated interface
var _ internalapi.ServerInterface = (*InternalServer)(nil)

// RegisterInternalRoutes registers internal endpoints
func RegisterInternalRoutes(router internalapi.EchoRouter, server *InternalServer) {
	// Use the generated handler registration
	internalapi.RegisterHandlers(router, server)
}

// InternalPing responds to internal health checks from other services
func (s *InternalServer) InternalPing(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

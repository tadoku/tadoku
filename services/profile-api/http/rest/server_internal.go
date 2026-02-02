package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/serviceauth"
	"github.com/tadoku/tadoku/services/profile-api/http/rest/openapi/internalapi"
)

// InternalServer handles internal service-to-service endpoints
type InternalServer struct{}

// NewInternalServer creates a new internal server
func NewInternalServer() *InternalServer {
	return &InternalServer{}
}

// Ensure InternalServer implements the generated interface
var _ internalapi.ServerInterface = (*InternalServer)(nil)

// RegisterInternalRoutes registers internal endpoints with service auth middleware
func RegisterInternalRoutes(e *echo.Echo, server *InternalServer, validator *serviceauth.TokenValidator) {
	internal := e.Group("")
	internal.Use(serviceauth.ServiceAuth(validator))

	// Use the generated handler registration
	internalapi.RegisterHandlers(internal, server)
}

// InternalPing responds to internal health checks from other services
func (s *InternalServer) InternalPing(c echo.Context) error {
	caller := serviceauth.GetCallingService(c)
	return c.JSON(http.StatusOK, internalapi.InternalPingResult{
		Status: "ok",
		Caller: caller,
	})
}

package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/serviceauth"
)

// InternalServer handles internal service-to-service endpoints
type InternalServer struct{}

// NewInternalServer creates a new internal server
func NewInternalServer() *InternalServer {
	return &InternalServer{}
}

// RegisterInternalRoutes registers internal endpoints with service auth middleware
func RegisterInternalRoutes(e *echo.Echo, server *InternalServer, validator *serviceauth.TokenValidator) {
	internal := e.Group("/internal/v1")
	internal.Use(serviceauth.ServiceAuth(validator))

	internal.GET("/ping", server.InternalPing)
}

// InternalPing responds to internal health checks from other services
func (s *InternalServer) InternalPing(c echo.Context) error {
	caller := serviceauth.GetCallingService(c)
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"caller": caller,
	})
}

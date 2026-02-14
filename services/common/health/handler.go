package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const checkTimeout = 2 * time.Second

// LivezHandler returns 200 OK if the process is alive.
func LivezHandler(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// ReadyzHandler checks all registered HealthCheckers and returns
// 200 if all pass, 503 if any fail.
func ReadyzHandler(checkers []HealthChecker) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), checkTimeout)
		defer cancel()

		response := ReadyzResponse{Status: "ready"}
		allHealthy := true

		for _, checker := range checkers {
			result := CheckResult{Name: checker.Name(), Status: "up"}
			if err := checker.Check(ctx); err != nil {
				result.Status = "down"
				result.Error = err.Error()
				allHealthy = false
			}
			response.Checks = append(response.Checks, result)
		}

		if !allHealthy {
			response.Status = "not_ready"
			return c.JSON(http.StatusServiceUnavailable, response)
		}

		return c.JSON(http.StatusOK, response)
	}
}

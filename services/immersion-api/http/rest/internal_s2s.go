package rest

import (
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/client/s2s"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

const defaultProfileAPIURL = "http://profile-api"

// RegisterInternalS2SRoutes registers internal routes that use S2S authentication.
func RegisterInternalS2SRoutes(e *echo.Echo, s2sClient *s2s.Client, profileAPIURL string, httpClient *http.Client) {
	if s2sClient == nil {
		return
	}

	if profileAPIURL == "" {
		profileAPIURL = defaultProfileAPIURL
	}
	profileAPIURL = strings.TrimRight(profileAPIURL, "/")

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	e.GET("/internal/v1/profile/ping", func(c echo.Context) error {
		if commondomain.ParseServiceIdentity(c.Request().Context()) == nil {
			return c.NoContent(http.StatusForbidden)
		}

		token, err := s2sClient.GetToken("profile-api")
		if err != nil {
			return c.String(http.StatusBadGateway, "failed to get s2s token")
		}

		req, err := http.NewRequest("GET", profileAPIURL+"/internal/v1/ping", nil)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := httpClient.Do(req)
		if err != nil {
			return c.String(http.StatusBadGateway, "profile-api request failed")
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}

		body, _ := io.ReadAll(resp.Body)
		return c.Blob(resp.StatusCode, contentType, body)
	})
}

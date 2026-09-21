package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRestoreJSONCharset_JSONAndHTTPError(t *testing.T) {
	e := echo.New()
	RestoreJSONCharset(e)
	e.GET("/ok", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/err", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "missing or malformed jwt")
	})

	for _, test := range []struct {
		path string
		want int
	}{
		{path: "/ok", want: http.StatusOK},
		{path: "/err", want: http.StatusBadRequest},
	} {
		response := httptest.NewRecorder()
		e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != test.want {
			t.Errorf("%s: status=%d, want %d", test.path, response.Code, test.want)
		}
		if got := response.Header().Get(echo.HeaderContentType); got != echo.MIMEApplicationJSONCharsetUTF8 {
			t.Errorf("%s: Content-Type=%q, want %q", test.path, got, echo.MIMEApplicationJSONCharsetUTF8)
		}
	}
}

package middleware

import "github.com/labstack/echo/v4"

// RestoreJSONCharset keeps Echo JSON responses on
// application/json; charset=UTF-8 after Echo v4.15 dropped the charset
// from MIMEApplicationJSON.
func RestoreJSONCharset(e *echo.Echo) {
	inner := e.JSONSerializer
	if inner == nil {
		inner = echo.DefaultJSONSerializer{}
	}
	e.JSONSerializer = jsonCharsetSerializer{inner: inner}
}

type jsonCharsetSerializer struct {
	inner echo.JSONSerializer
}

func (s jsonCharsetSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	header := c.Response().Header()
	ct := header.Get(echo.HeaderContentType)
	if ct == "" || ct == echo.MIMEApplicationJSON {
		header.Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	}
	return s.inner.Serialize(c, i, indent)
}

func (s jsonCharsetSerializer) Deserialize(c echo.Context, i interface{}) error {
	return s.inner.Deserialize(c, i)
}

package rest

import (
	"github.com/labstack/echo/v4"
	commonhttperr "github.com/tadoku/tadoku/services/common/http/httperr"
)

func handleCommonErrors(ctx echo.Context, err error) (bool, error) {
	if code, ok := commonhttperr.StatusCode(err); ok {
		return true, ctx.NoContent(code)
	}
	return false, nil
}

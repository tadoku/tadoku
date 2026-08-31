package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	commonhttperr "github.com/tadoku/tadoku/services/common/http/httperr"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func handleCommonErrors(ctx echo.Context, err error) (bool, error) {
	if errors.Is(err, domain.ErrAccountDeletionInProgress) {
		return true, ctx.JSON(http.StatusConflict, map[string]string{
			"error": "account_deletion_in_progress",
		})
	}
	if code, ok := commonhttperr.StatusCode(err); ok {
		return true, ctx.NoContent(code)
	}
	return false, nil
}

package rest

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi/internalapi"
)

type accountDeletionLocker interface {
	Execute(context.Context, *domain.AccountDeletionLockRequest) error
}

type InternalServer struct {
	accountDeletionLock accountDeletionLocker
}

func NewInternalServer(accountDeletionLock accountDeletionLocker) *InternalServer {
	return &InternalServer{accountDeletionLock: accountDeletionLock}
}

var _ internalapi.ServerInterface = (*InternalServer)(nil)

func RegisterInternalRoutes(router internalapi.EchoRouter, server *InternalServer) {
	internalapi.RegisterHandlers(router, server)
}

func (s *InternalServer) InternalAccountDeletionLock(ctx echo.Context) error {
	var body internalapi.InternalAccountDeletionLockJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.accountDeletionLock.Execute(ctx.Request().Context(), &domain.AccountDeletionLockRequest{
		UserID:    body.UserId,
		RequestID: body.RequestId,
	})
	if err != nil {
		if handled, responseErr := handleCommonErrors(ctx, err); handled {
			return responseErr
		}
		ctx.Echo().Logger.Error("could not lock account for deletion: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

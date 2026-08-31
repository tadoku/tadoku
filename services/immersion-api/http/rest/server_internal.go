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

type accountDeletionScrubber interface {
	Execute(context.Context, *domain.AccountDeletionScrubRequest) error
}

type InternalServer struct {
	accountDeletionLock  accountDeletionLocker
	accountDeletionScrub accountDeletionScrubber
}

func NewInternalServer(accountDeletionLock accountDeletionLocker, accountDeletionScrub accountDeletionScrubber) *InternalServer {
	return &InternalServer{
		accountDeletionLock:  accountDeletionLock,
		accountDeletionScrub: accountDeletionScrub,
	}
}

func (s *InternalServer) InternalAccountDeletionScrub(ctx echo.Context) error {
	var body internalapi.InternalAccountDeletionScrubJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.accountDeletionScrub.Execute(ctx.Request().Context(), &domain.AccountDeletionScrubRequest{
		UserID:    body.UserId,
		RequestID: body.RequestId,
	})
	if err != nil {
		if handled, responseErr := handleCommonErrors(ctx, err); handled {
			return responseErr
		}
		ctx.Echo().Logger.Error("could not scrub account: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
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

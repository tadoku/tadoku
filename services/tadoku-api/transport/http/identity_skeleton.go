package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (*server) IdentityRecoveryCreate(
	ctx context.Context,
	request openapi.IdentityRecoveryCreateRequestObject,
) (openapi.IdentityRecoveryCreateResponseObject, error) {
	if request.Body != nil {
		if err := app.ValidateIdentityReturnTo(request.Body.ReturnTo); err != nil {
			return nil, err
		}
	}
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityRecoveryConsume(
	context.Context,
	openapi.IdentityRecoveryConsumeRequestObject,
) (openapi.IdentityRecoveryConsumeResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityProfileUpdate(
	context.Context,
	openapi.IdentityProfileUpdateRequestObject,
) (openapi.IdentityProfileUpdateResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityPasswordUpdate(
	context.Context,
	openapi.IdentityPasswordUpdateRequestObject,
) (openapi.IdentityPasswordUpdateResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) EmailList(
	context.Context,
	openapi.EmailListRequestObject,
) (openapi.EmailListResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) EmailGet(
	context.Context,
	openapi.EmailGetRequestObject,
) (openapi.EmailGetResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

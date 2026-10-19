package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (*server) IdentityRegistrationCreate(
	context.Context,
	openapi.IdentityRegistrationCreateRequestObject,
) (openapi.IdentityRegistrationCreateResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityLoginCreate(
	context.Context,
	openapi.IdentityLoginCreateRequestObject,
) (openapi.IdentityLoginCreateResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityLogout(
	context.Context,
	openapi.IdentityLogoutRequestObject,
) (openapi.IdentityLogoutResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentitySessionGet(
	context.Context,
	openapi.IdentitySessionGetRequestObject,
) (openapi.IdentitySessionGetResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

func (*server) IdentityRecoveryCreate(
	context.Context,
	openapi.IdentityRecoveryCreateRequestObject,
) (openapi.IdentityRecoveryCreateResponseObject, error) {
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

func (*server) IdentityProxySessionCheck(
	context.Context,
	callbackopenapi.IdentityProxySessionCheckRequestObject,
) (callbackopenapi.IdentityProxySessionCheckResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

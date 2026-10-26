package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *server) IdentityRecoveryCreate(
	ctx context.Context,
	request openapi.IdentityRecoveryCreateRequestObject,
) (openapi.IdentityRecoveryCreateResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	command := api.IdentityRecovery{IP: metadata.ip}
	if request.Body != nil {
		command.Email, command.ReturnTo = string(request.Body.Email), request.Body.ReturnTo
	}
	if err := api.ValidateIdentityReturnTo(command.ReturnTo); err != nil {
		return nil, err
	}
	if err := s.application.IdentityRequestRecovery(ctx, command); err != nil {
		s.logOperationError(ctx, "request account recovery", err)
		return nil, err
	}
	return openapi.IdentityRecoveryCreate202Response{}, nil
}

func (s *server) IdentityRecoveryConsume(
	ctx context.Context,
	request openapi.IdentityRecoveryConsumeRequestObject,
) (openapi.IdentityRecoveryConsumeResponseObject, error) {
	var returnTo *string
	if request.Params.ReturnTo != nil {
		value := *request.Params.ReturnTo
		returnTo = &value
	}
	result, err := s.application.IdentityConsumeRecovery(ctx, api.IdentityRecoveryToken(request.Token), returnTo)
	if err != nil {
		s.logOperationError(ctx, "consume account recovery", err)
		return nil, err
	}
	cookie := identityCookie(result.Login.Cookie)
	return openapi.IdentityRecoveryConsume303Response{
		Headers: openapi.IdentityRecoveryConsume303ResponseHeaders{Location: &result.RedirectTo, SetCookie: &cookie},
	}, nil
}

func (s *server) IdentityProfileUpdate(
	ctx context.Context,
	request openapi.IdentityProfileUpdateRequestObject,
) (openapi.IdentityProfileUpdateResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	var command api.IdentityProfileSettings
	if request.Body != nil {
		command.DisplayName = request.Body.DisplayName
		if request.Body.Email != nil {
			email := string(*request.Body.Email)
			command.Email = &email
		}
	}
	result, err := s.application.IdentityUpdateProfile(ctx, metadata.cookies, command)
	if err != nil {
		s.logOperationError(ctx, "update account profile", err)
		return nil, err
	}
	return openapi.IdentityProfileUpdate200JSONResponse{
		Id:              result.Account.ID,
		Email:           accountEmail(result.Account.Email.String()),
		DisplayName:     result.Account.DisplayName.String(),
		CreatedAt:       result.Account.CreatedAt,
		AuthenticatedAt: result.AuthenticatedAt,
	}, nil
}

func (s *server) IdentityPasswordUpdate(
	ctx context.Context,
	request openapi.IdentityPasswordUpdateRequestObject,
) (openapi.IdentityPasswordUpdateResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	password := ""
	if request.Body != nil {
		password = request.Body.Password
	}
	if err := s.application.IdentityUpdatePassword(ctx, metadata.cookies, password); err != nil {
		s.logOperationError(ctx, "update account password", err)
		return nil, err
	}
	return openapi.IdentityPasswordUpdate204Response{}, nil
}

func (s *server) IdentityAccountDelete(
	ctx context.Context,
	request openapi.IdentityAccountDeleteRequestObject,
) (openapi.IdentityAccountDeleteResponseObject, error) {
	return nil, errx.NewNotImplementedError()
}

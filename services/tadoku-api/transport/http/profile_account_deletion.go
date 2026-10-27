package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ProfileAccountDeletionCreate(
	ctx context.Context,
	request openapi.ProfileAccountDeletionCreateRequestObject,
) (openapi.ProfileAccountDeletionCreateResponseObject, error) {
	parameters := api.AccountDeletionParameters{
		UserID: request.Id,
		Email:  request.Body.Email,
		Reason: request.Body.Reason,
	}

	if err := s.application.RequestAccountDeletion(ctx, parameters); err != nil {
		s.logOperationError(ctx, "request account deletion", err)
		return nil, err
	}

	return openapi.ProfileAccountDeletionCreate202Response{}, nil
}

package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLogContestRegistrationUpdate(ctx context.Context, request openapi.ImmersionLogContestRegistrationUpdateRequestObject) (openapi.ImmersionLogContestRegistrationUpdateResponseObject, error) {
	updated, err := s.application.UpdateLogContestRegistrations(ctx, request.Id, request.Body.RegistrationIds)
	if err != nil {
		s.logOperationError(ctx, "update log contest registrations", err)
		return nil, err
	}

	return openapi.ImmersionLogContestRegistrationUpdate200JSONResponse(logDetailResponse(*updated)), nil
}

func (s *server) ImmersionLogDeleteByID(ctx context.Context, request openapi.ImmersionLogDeleteByIDRequestObject) (openapi.ImmersionLogDeleteByIDResponseObject, error) {
	if err := s.application.DeleteLog(ctx, request.Id); err != nil {
		s.logOperationError(ctx, "delete log", err)
		return nil, err
	}

	return openapi.ImmersionLogDeleteByID200Response{}, nil
}

func (s *server) ImmersionContestModerationDetachLog(ctx context.Context, request openapi.ImmersionContestModerationDetachLogRequestObject) (openapi.ImmersionContestModerationDetachLogResponseObject, error) {
	if err := s.application.DetachContestLog(ctx, request.Id, request.LogId, request.Body.Reason); err != nil {
		s.logOperationError(ctx, "detach contest log", err)
		return nil, err
	}

	return openapi.ImmersionContestModerationDetachLog200Response{}, nil
}

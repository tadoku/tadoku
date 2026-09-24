package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLogCreate(ctx context.Context, request openapi.ImmersionLogCreateRequestObject) (openapi.ImmersionLogCreateResponseObject, error) {
	body := request.Body
	parameters := app.LogCreateParameters{}
	if body != nil {
		parameters = app.LogCreateParameters{
			UnitID:          body.UnitId,
			UnitKey:         body.UnitKey,
			ActivityID:      body.ActivityId,
			LanguageCode:    body.LanguageCode,
			Amount:          body.Amount,
			DurationSeconds: body.DurationSeconds,
			Tags:            body.Tags,
			Description:     body.Description,
		}
		if body.RegistrationIds != nil {
			parameters.RegistrationIDs = *body.RegistrationIds
		}
	}
	created, err := s.application.CreateLog(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "create log", err)
		return nil, err
	}
	return openapi.ImmersionLogCreate200JSONResponse(logDetailResponse(*created)), nil
}

func (s *server) ImmersionLogUpdate(ctx context.Context, request openapi.ImmersionLogUpdateRequestObject) (openapi.ImmersionLogUpdateResponseObject, error) {
	body := request.Body
	updated, err := s.application.UpdateLog(ctx, app.LogUpdateParameters{
		ID:              request.Id,
		UnitID:          body.UnitId,
		UnitKey:         body.UnitKey,
		Amount:          body.Amount,
		DurationSeconds: body.DurationSeconds,
		Tags:            body.Tags,
		Description:     body.Description,
	})
	if err != nil {
		s.logOperationError(ctx, "update log", err)
		return nil, err
	}
	return openapi.ImmersionLogUpdate200JSONResponse(logDetailResponse(*updated)), nil
}

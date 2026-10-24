package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestListEvents(
	ctx context.Context,
	request openapi.ImmersionContestListEventsRequestObject,
) (openapi.ImmersionContestListEventsResponseObject, error) {
	cursor := "0"
	if request.Params.Cursor != nil {
		cursor = *request.Params.Cursor
	}

	limit := 100
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	page, err := s.application.ListContestEvents(ctx, request.Id, cursor, limit)
	if err != nil {
		s.logOperationError(ctx, "list contest events", err)
		return nil, err
	}

	events := make([]openapi.ImmersionContestEvent, 0, len(page.Events))
	for _, event := range page.Events {
		events = append(events, openapi.ImmersionContestEvent{
			Id:    event.ID,
			LogId: event.LogID,
			Event: openapi.ImmersionContestEventEvent(event.Type),
		})
	}

	return openapi.ImmersionContestListEvents200JSONResponse{
		Body: openapi.ImmersionContestEventPage{
			Events:     events,
			NextCursor: page.NextCursor,
			HasMore:    page.HasMore,
		},
		Headers: openapi.ImmersionContestListEvents200ResponseHeaders{
			CacheControl: pointer(privateNoStore),
		},
	}, nil
}

func (s *server) ImmersionContestFindLog(
	ctx context.Context,
	request openapi.ImmersionContestFindLogRequestObject,
) (openapi.ImmersionContestFindLogResponseObject, error) {
	log, err := s.application.FindContestLog(ctx, request.Id, request.LogId)
	if err != nil {
		s.logOperationError(ctx, "find contest log", err)
		return nil, err
	}

	return openapi.ImmersionContestFindLog200JSONResponse{
		Body: logResponse(*log),
		Headers: openapi.ImmersionContestFindLog200ResponseHeaders{
			CacheControl: pointer(privateNoStore),
		},
	}, nil
}

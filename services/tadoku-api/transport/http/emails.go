package http

import (
	"context"
	"time"

	"github.com/oapi-codegen/nullable"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) EmailList(
	ctx context.Context,
	request openapi.EmailListRequestObject,
) (openapi.EmailListResponseObject, error) {
	filter := app.EmailFilter{Since: request.Params.Since}
	if request.Params.Recipient != nil {
		filter.Recipient = string(*request.Params.Recipient)
	}
	if request.Params.Template != nil {
		filter.Template = app.EmailTemplate(*request.Params.Template)
	}
	if request.Params.Status != nil {
		filter.Status = app.EmailStatus(*request.Params.Status)
	}
	if request.Params.Limit != nil {
		filter.Limit = *request.Params.Limit
	}
	if request.Params.Cursor != nil {
		filter.Cursor = *request.Params.Cursor
	}
	page, err := s.application.ListEmails(ctx, filter)
	if err != nil {
		return nil, err
	}
	response := openapi.EmailList200JSONResponse{
		Emails:     make([]openapi.EmailMessage, 0, len(page.Emails)),
		NextCursor: nullable.NewNullNullable[string](),
	}
	if page.NextCursor != "" {
		response.NextCursor = nullable.NewNullableWithValue(page.NextCursor)
	}
	for _, row := range page.Emails {
		response.Emails = append(response.Emails, emailResponse(row))
	}
	return response, nil
}

func (s *server) EmailGet(
	ctx context.Context,
	request openapi.EmailGetRequestObject,
) (openapi.EmailGetResponseObject, error) {
	row, err := s.application.FindEmail(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return openapi.EmailGet200JSONResponse(emailResponse(row)), nil
}

func emailResponse(row app.Email) openapi.EmailMessage {
	result := openapi.EmailMessage{
		Id:                row.ID,
		Tenant:            row.Tenant.String(),
		Template:          string(row.Template),
		Recipient:         openapitypes.Email(row.Recipient),
		Subject:           row.Subject,
		Body:              row.Body,
		Params:            row.Params,
		CreatedAt:         row.CreatedAt,
		SentAt:            nullable.NewNullNullable[time.Time](),
		ProviderMessageId: nullable.NewNullNullable[string](),
		Error:             nullable.NewNullNullable[string](),
	}
	switch delivery := row.Delivery.(type) {
	case app.EmailPending:
		result.Status = openapi.EmailStatus("pending")
	case app.EmailHeld:
		result.Status = openapi.EmailStatus("held")
	case app.EmailSent:
		result.Status = openapi.EmailStatus("sent")
		result.SentAt = nullable.NewNullableWithValue(delivery.At)
		result.ProviderMessageId = nullable.NewNullableWithValue(delivery.ProviderMessageID)
	case app.EmailFailed:
		result.Status = openapi.EmailStatus("failed")
		result.Error = nullable.NewNullableWithValue(delivery.Error)
	}
	return result
}

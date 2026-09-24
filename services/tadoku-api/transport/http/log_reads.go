package http

import (
	"context"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLogFindByID(ctx context.Context, request openapi.ImmersionLogFindByIDRequestObject) (openapi.ImmersionLogFindByIDResponseObject, error) {
	log, err := s.application.FindLog(ctx, request.Id)
	if err != nil {
		s.logOperationError(ctx, "find log", err)
		return nil, err
	}

	response := logDetailResponse(*log)
	return openapi.ImmersionLogFindByID200JSONResponse(response), nil
}

func logDetailResponse(log app.Log) openapi.ImmersionLog {
	response := logResponse(log)
	response.UnitId = log.UnitID
	if log.UnitKey != "" {
		response.UnitKey = &log.UnitKey
	}
	refs := make([]openapi.ImmersionContestRegistrationReference, 0, len(log.Registrations))
	for _, ref := range log.Registrations {
		refs = append(refs, openapi.ImmersionContestRegistrationReference{
			ContestId:            ref.ContestID,
			ContestEnd:           openapiTypes.Date{Time: ref.ContestEnd},
			RegistrationId:       ref.RegistrationID,
			Title:                ref.Title,
			OwnerUserDisplayName: &ref.OwnerUserDisplayName,
			Official:             &ref.Official,
			Score:                &ref.Score,
		})
	}
	response.Registrations = &refs
	return response
}

func (s *server) ImmersionProfileListLogs(ctx context.Context, request openapi.ImmersionProfileListLogsRequestObject) (openapi.ImmersionProfileListLogsResponseObject, error) {
	parameters := app.LogListParameters{}
	parameters.UserID = &request.UserId
	if request.Params.PageSize != nil {
		parameters.PageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		parameters.Page = *request.Params.Page
	}
	if request.Params.IncludeDeleted != nil {
		parameters.IncludeDeleted = *request.Params.IncludeDeleted
	}

	result, err := s.application.ListUserLogs(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "list user logs", err)
		return nil, err
	}

	return openapi.ImmersionProfileListLogs200JSONResponse(logListResponse(result)), nil
}

func (s *server) ImmersionContestListLogs(ctx context.Context, request openapi.ImmersionContestListLogsRequestObject) (openapi.ImmersionContestListLogsResponseObject, error) {
	parameters := app.LogListParameters{}
	parameters.ContestID = request.Id
	parameters.UserID = request.Params.UserId
	if request.Params.PageSize != nil {
		parameters.PageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		parameters.Page = *request.Params.Page
	}
	if request.Params.IncludeDeleted != nil {
		parameters.IncludeDeleted = *request.Params.IncludeDeleted
	}

	result, err := s.application.ListContestLogs(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "list contest logs", err)
		return nil, err
	}

	return openapi.ImmersionContestListLogs200JSONResponse(logListResponse(result)), nil
}

func logListResponse(result *app.LogList) openapi.ImmersionLogs {
	response := openapi.ImmersionLogs{
		Logs:          make([]openapi.ImmersionLog, 0, len(result.Logs)),
		TotalSize:     result.TotalSize,
		NextPageToken: result.NextPageToken,
	}
	for _, log := range result.Logs {
		response.Logs = append(response.Logs, logResponse(log))
	}
	return response
}

func logResponse(log app.Log) openapi.ImmersionLog {
	inputType := openapi.ImmersionActivityInputType(log.Activity.InputType)
	return openapi.ImmersionLog{
		Id: log.ID,
		Activity: openapi.ImmersionActivity{
			Id:        log.Activity.ID,
			Name:      log.Activity.Name,
			InputType: &inputType,
		},
		Language: openapi.ImmersionLanguage{
			Code: log.LanguageCode,
			Name: log.LanguageName,
		},
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.Score,
		DurationSeconds: log.DurationSeconds,
		Tags:            log.Tags,
		UnitName:        log.UnitName,
		UserId:          log.UserID,
		UserDisplayName: log.UserDisplayName,
		CreatedAt:       log.CreatedAt,
		Deleted:         log.Deleted,
		Description:     log.Description,
	}
}

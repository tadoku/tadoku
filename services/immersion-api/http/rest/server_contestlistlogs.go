package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Lists the logs attached to a contest
// (GET /contests/{id}/logs)
func (s *Server) ContestListLogs(ctx echo.Context, id types.UUID, params openapi.ContestListLogsParams) error {
	req := &domain.LogListForContestRequest{
		UserID:         uuid.NullUUID{},
		ContestID:      id,
		IncludeDeleted: false,
		PageSize:       0,
		Page:           0,
	}

	if params.PageSize != nil {
		req.PageSize = *params.PageSize
	}
	if params.Page != nil {
		req.Page = *params.Page
	}
	if params.IncludeDeleted != nil {
		req.IncludeDeleted = *params.IncludeDeleted
	}
	if params.UserId != nil {
		req.UserID = uuid.NullUUID{
			UUID:  *params.UserId,
			Valid: true,
		}
	}

	list, err := s.logListForContest.Execute(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return ctx.NoContent(http.StatusForbidden)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.Logs{
		Logs:          make([]openapi.Log, len(list.Logs)),
		NextPageToken: list.NextPageToken,
		TotalSize:     list.TotalSize,
	}

	for i, it := range list.Logs {
		res.Logs[i] = openapi.Log{
			Id: it.ID,
			Activity: openapi.Activity{
				Id:   int32(it.ActivityID),
				Name: it.ActivityName,
			},
			Language: openapi.Language{
				Code: it.LanguageCode,
				Name: it.LanguageName,
			},
			Amount:          it.Amount,
			Modifier:        it.Modifier,
			Score:           it.EffectiveScore(),
			Tags:            it.Tags,
			UnitName:        it.UnitName,
			DurationSeconds: intPtrFromInt32Ptr(it.DurationSeconds),
			UserId:          it.UserID,
			UserDisplayName: it.UserDisplayName,
			CreatedAt:       it.CreatedAt,
			Deleted:         it.Deleted,
			Description:     it.Description,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

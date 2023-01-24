package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a log by id
// (GET /logs/{id})
func (s *Server) LogFindByID(ctx echo.Context, id types.UUID) error {
	log, err := s.queryService.FindLogByID(ctx.Request().Context(), &query.FindLogByIDRequest{
		ID: id,
	})
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	refs := make([]openapi.ContestRegistrationReference, len(log.Registrations))
	for i, it := range log.Registrations {
		refs[i] = openapi.ContestRegistrationReference{
			ContestId:      it.ContestID,
			ContestEnd:     types.Date{Time: it.ContestEnd},
			RegistrationId: it.RegistrationID,
			Title:          it.Title,
		}
	}

	res := openapi.Log{
		Id: log.ID,
		Activity: openapi.Activity{
			Id:   int32(log.ActivityID),
			Name: log.ActivityName,
		},
		Language: openapi.Language{
			Code: log.LanguageCode,
			Name: log.LanguageName,
		},
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.Score,
		Tags:            log.Tags,
		UnitName:        log.UnitName,
		UserId:          log.UserID,
		UserDisplayName: log.UserDisplayName,
		CreatedAt:       log.CreatedAt,
		Deleted:         log.Deleted,
		Description:     log.Description,
		Registrations:   &refs,
	}

	return ctx.JSON(http.StatusOK, res)
}

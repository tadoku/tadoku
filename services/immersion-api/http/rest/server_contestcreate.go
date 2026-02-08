package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Creates a new contest
// (POST /contests)
func (s *Server) ContestCreate(ctx echo.Context) error {
	var req openapi.ContestCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	contest, err := s.contestCreate.Execute(ctx.Request().Context(), &domain.ContestCreateRequest{
		Official:                req.Official,
		Private:                 req.Private,
		ContestStart:            req.ContestStart.Time,
		ContestEnd:              req.ContestEnd.Time,
		RegistrationEnd:         req.RegistrationEnd.Time,
		Title:                   req.Title,
		Description:             req.Description,
		LanguageCodeAllowList:   req.LanguageCodeAllowList,
		ActivityTypeIDAllowList: req.ActivityTypeIdAllowList,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		if errors.Is(err, domain.ErrInvalidContest) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.Contest{
		Id:                      &contest.ID,
		ContestStart:            types.Date{Time: contest.ContestStart},
		ContestEnd:              types.Date{Time: contest.ContestEnd},
		RegistrationEnd:         types.Date{Time: contest.RegistrationEnd},
		Title:                   contest.Title,
		Description:             contest.Description,
		OwnerUserId:             &contest.OwnerUserID,
		OwnerUserDisplayName:    &contest.OwnerUserDisplayName,
		Official:                contest.Official,
		Private:                 contest.Private,
		LanguageCodeAllowList:   contest.LanguageCodeAllowList,
		ActivityTypeIdAllowList: contest.ActivityTypeIDAllowList,
		CreatedAt:               &contest.CreatedAt,
		UpdatedAt:               &contest.UpdatedAt,
	})
}

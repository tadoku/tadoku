package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a contest by id
// (GET /contests/{id})
func (s *Server) ContestFindByID(ctx echo.Context, id types.UUID) error {
	contest, err := s.contestFind.Execute(ctx.Request().Context(), &domain.ContestFindRequest{
		ID: id,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	langs := make([]openapi.Language, len(contest.AllowedLanguages))
	for i, it := range contest.AllowedLanguages {
		langs[i] = openapi.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	if len(langs) == 0 {
		langs = nil
	}

	acts := make([]openapi.Activity, len(contest.AllowedActivities))
	for i, it := range contest.AllowedActivities {
		acts[i] = openapi.Activity{
			Id:        it.ID,
			Name:      it.Name,
			InputType: openapi.ActivityInputType(it.InputType),
		}
	}

	return ctx.JSON(http.StatusOK, openapi.ContestView{
		Id:                   &contest.ID,
		ContestStart:         types.Date{Time: contest.ContestStart},
		ContestEnd:           types.Date{Time: contest.ContestEnd},
		RegistrationEnd:      types.Date{Time: contest.RegistrationEnd},
		Title:                contest.Title,
		Description:          contest.Description,
		OwnerUserId:          &contest.OwnerUserID,
		OwnerUserDisplayName: &contest.OwnerUserDisplayName,
		Official:             contest.Official,
		Private:              contest.Private,
		AllowedLanguages:     langs,
		AllowedActivities:    acts,
		CreatedAt:            &contest.CreatedAt,
		UpdatedAt:            &contest.UpdatedAt,
		Deleted:              &contest.Deleted,
	})
}

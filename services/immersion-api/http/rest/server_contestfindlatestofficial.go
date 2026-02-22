package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the latest official contest
// (GET /contests/latest-official)
func (s *Server) ContestFindLatestOfficial(ctx echo.Context) error {
	contest, err := s.contestFindLatestOfficial.Execute(ctx.Request().Context())
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

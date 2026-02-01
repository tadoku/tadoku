package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Lists all the contests, paginated
// (GET /contests)
func (s *Server) ContestList(ctx echo.Context, params openapi.ContestListParams) error {
	pageSize := 0
	page := 0
	includeDeleted := false
	officialOnly := true
	userID := uuid.NullUUID{}

	if params.PageSize != nil {
		pageSize = *params.PageSize
	}
	if params.Page != nil {
		page = *params.Page
	}
	if params.IncludeDeleted != nil {
		includeDeleted = *params.IncludeDeleted
	}
	if params.Official != nil {
		officialOnly = *params.Official
	}
	if params.UserId != nil {
		userID = uuid.NullUUID{
			UUID:  *params.UserId,
			Valid: true,
		}
	}

	list, err := s.contestList.Execute(ctx.Request().Context(), &domain.ContestListRequest{
		UserID:         userID,
		OfficialOnly:   officialOnly,
		IncludeDeleted: includeDeleted,
		PageSize:       pageSize,
		Page:           page,
	})
	if err != nil {
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.Contests{
		Contests:      make([]openapi.Contest, len(list.Contests)),
		NextPageToken: list.NextPageToken,
		TotalSize:     list.TotalSize,
	}

	for i, contest := range list.Contests {
		res.Contests[i] = openapi.Contest{
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
			Deleted:                 &contest.Deleted,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

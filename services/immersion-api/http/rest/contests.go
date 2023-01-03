package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain/contestcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// COMMANDS

// Creates a new contest
// (POST /contests)
func (s *Server) ContestCreate(ctx echo.Context) error {
	var req openapi.ContestCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	contest, err := s.contestCommandService.CreateContest(ctx.Request().Context(), &contestcommand.ContestCreateRequest{
		Official:                req.Official,
		Private:                 req.Private,
		ContestStart:            req.ContestStart.Time,
		ContestEnd:              req.ContestEnd.Time,
		RegistrationEnd:         req.RegistrationEnd.Time,
		Description:             req.Description,
		LanguageCodeAllowList:   req.LanguageCodeAllowList,
		ActivityTypeIDAllowList: req.ActivityTypeIdAllowList,
	})
	if err != nil {
		if errors.Is(err, contestcommand.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, contestcommand.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, contestcommand.ErrInvalidContest) {
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

// QUERIES

// Fetches the configuration options for a new contest
// (GET /contests/configuration-options)
func (s *Server) ContestGetConfigurations(ctx echo.Context) error {
	opts, err := s.contestQueryService.FetchContestConfigurationOptions(ctx.Request().Context())
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.ContestConfigurationOptions{
		Activities:             make([]openapi.Activity, len(opts.Activities)),
		Languages:              make([]openapi.Language, len(opts.Languages)),
		CanCreateOfficialRound: opts.CanCreateOfficialRound,
	}

	for i, a := range opts.Activities {
		a := a
		res.Activities[i] = openapi.Activity{
			Id:      a.ID,
			Name:    a.Name,
			Default: &a.Default,
		}
	}

	for i, l := range opts.Languages {
		res.Languages[i] = openapi.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches a contest by id
// (GET /contests/{id})
func (s *Server) ContestFindByID(ctx echo.Context, id types.UUID) error {
	contest, err := s.contestQueryService.FindByID(ctx.Request().Context(), &contestquery.FindByIDRequest{
		ID: id,
	})
	if err != nil {
		if errors.Is(err, contestquery.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
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

	acts := make([]openapi.Activity, len(contest.AllowedActivities))
	for i, it := range contest.AllowedActivities {
		acts[i] = openapi.Activity{
			Id:   it.ID,
			Name: it.Name,
		}
	}

	return ctx.JSON(http.StatusOK, openapi.ContestView{
		Id:                   &contest.ID,
		ContestStart:         types.Date{Time: contest.ContestStart},
		ContestEnd:           types.Date{Time: contest.ContestEnd},
		RegistrationEnd:      types.Date{Time: contest.RegistrationEnd},
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

	list, err := s.contestQueryService.ListContests(ctx.Request().Context(), &contestquery.ContestListRequest{
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
		contest := contest
		res.Contests[i] = openapi.Contest{
			Id:                      &contest.ID,
			ContestStart:            types.Date{Time: contest.ContestStart},
			ContestEnd:              types.Date{Time: contest.ContestEnd},
			RegistrationEnd:         types.Date{Time: contest.RegistrationEnd},
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

// Fetches a contest registration if it exists
// (GET /contests/{id}/registration)
func (s *Server) ContestFindRegistration(ctx echo.Context, id types.UUID) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

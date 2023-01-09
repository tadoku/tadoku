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

// Creates or updates a registration for a contest
// (POST /contests/{id}/registration)
func (s *Server) ContestRegistrationUpsert(ctx echo.Context, id types.UUID) error {
	var req openapi.ContestRegistrationUpsertJSONBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.contestCommandService.UpsertContestRegistration(ctx.Request().Context(), &contestcommand.UpsertContestRegistrationRequest{
		ContestID:     id,
		LanguageCodes: req.LanguageCodes,
	})
	if err != nil {
		if errors.Is(err, contestquery.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, contestcommand.ErrInvalidContestRegistration) {
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
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

	if len(langs) == 0 {
		langs = nil
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
	reg, err := s.contestQueryService.FindRegistrationForUser(ctx.Request().Context(), &contestquery.FindRegistrationForUserRequest{
		ContestID: id,
	})
	if err != nil {
		if errors.Is(err, contestquery.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	langs := make([]openapi.Language, len(reg.Languages))
	for i, it := range reg.Languages {
		langs[i] = openapi.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	return ctx.JSON(http.StatusOK, openapi.ContestRegistration{
		Id:              &reg.ID,
		ContestId:       reg.ContestID,
		UserId:          reg.UserID,
		UserDisplayName: reg.UserDisplayName,
		Languages:       langs,
	})
}

// Fetches the leaderboard for a contest
// (GET /contests/{id}/leaderboard)
func (s *Server) ContestFetchLeaderboard(ctx echo.Context, id types.UUID, params openapi.ContestFetchLeaderboardParams) error {
	req := &contestquery.FetchContestLeaderboardRequest{
		ContestID:    id,
		LanguageCode: params.LanguageCode,
	}

	if params.PageSize != nil {
		req.PageSize = *params.PageSize
	}
	if params.Page != nil {
		req.Page = *params.Page
	}
	if params.ActivityId != nil {
		id := int32(*params.ActivityId)
		req.ActivityID = &id
	}

	leaderboard, err := s.contestQueryService.FetchContestLeaderboard(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, contestquery.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.Leaderboard{
		Entries:       make([]openapi.LeaderboardEntry, len(leaderboard.Entries)),
		NextPageToken: leaderboard.NextPageToken,
		TotalSize:     leaderboard.TotalSize,
	}

	for i, entry := range leaderboard.Entries {
		entry := entry
		res.Entries[i] = openapi.LeaderboardEntry{
			Rank:            entry.Rank,
			UserId:          entry.UserID,
			UserDisplayName: entry.UserDisplayName,
			Score:           int(entry.Score),
			IsTie:           entry.IsTie,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches all the ongoing contest registrations of the logged in user, always in a single page
// (GET /contests/configuration-options)
func (s *Server) ContestFindOngoingRegistrations(ctx echo.Context) error {
	regs, err := s.contestQueryService.FetchOngoingContestRegistrations(ctx.Request().Context(), &contestquery.FetchOngoingContestRegistrationsRequest{})
	if err != nil {
		if errors.Is(err, contestquery.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := &openapi.ContestRegistrations{
		TotalSize:     regs.TotalSize,
		NextPageToken: regs.NextPageToken,
		Registrations: make([]openapi.ContestRegistration, len(regs.Registrations)),
	}

	for i, r := range regs.Registrations {
		r := r
		contest := openapi.ContestView{
			Id:                &r.ContestID,
			ContestStart:      types.Date{Time: r.Contest.ContestStart},
			ContestEnd:        types.Date{Time: r.Contest.ContestEnd},
			RegistrationEnd:   types.Date{Time: r.Contest.RegistrationEnd},
			Description:       r.Contest.Description,
			Official:          r.Contest.Official,
			Private:           r.Contest.Private,
			AllowedLanguages:  []openapi.Language{},
			AllowedActivities: make([]openapi.Activity, len(r.Contest.AllowedActivities)),
		}

		for i, a := range r.Contest.AllowedActivities {
			contest.AllowedActivities[i] = openapi.Activity{
				Id:   a.ID,
				Name: a.Name,
			}
		}

		registration := openapi.ContestRegistration{
			ContestId:       r.ContestID,
			Id:              &r.ID,
			Languages:       make([]openapi.Language, len(r.Languages)),
			UserId:          r.UserID,
			UserDisplayName: r.UserDisplayName,
			Contest:         &contest,
		}

		for i, lang := range r.Languages {
			registration.Languages[i] = openapi.Language{
				Code: lang.Code,
				Name: lang.Name,
			}
		}

		res.Registrations[i] = registration
	}

	return ctx.JSON(http.StatusOK, res)
}

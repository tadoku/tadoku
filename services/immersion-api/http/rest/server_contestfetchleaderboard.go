package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the leaderboard for a contest
// (GET /contests/{id}/leaderboard)
func (s *Server) ContestFetchLeaderboard(ctx echo.Context, id types.UUID, params openapi.ContestFetchLeaderboardParams) error {
	req := &query.FetchContestLeaderboardRequest{
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

	leaderboard, err := s.queryService.FetchContestLeaderboard(ctx.Request().Context(), req)
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
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

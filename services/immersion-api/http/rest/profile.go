package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/profilequery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// COMMANDS

// QUERIES

// Fetches a profile of a user
// (GET /users/{userId}/profile)
func (s *Server) ProfileFindByUserID(ctx echo.Context, userId types.UUID) error {
	profile, err := s.profileQueryService.FetchUserProfile(ctx.Request().Context(), userId)
	if err != nil {
		if errors.Is(err, profilequery.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Errorf("could not fetch profile: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := &openapi.UserProfile{
		Id:          userId,
		DisplayName: profile.DisplayName,
		CreatedAt:   profile.CreatedAt,
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches a activity summary of a user for a given year
// (GET /users/{userId}/activity/{year})
func (s *Server) ProfileYearlyActivityByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.profileQueryService.YearlyActivityForUser(ctx.Request().Context(), &profilequery.YearlyActivityForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Echo().Logger.Errorf("could not fetch activity summary: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.UserActivityScore, len(summary.Scores))
	for i, it := range summary.Scores {
		scores[i] = openapi.UserActivityScore{
			Date:  types.Date{Time: it.Date},
			Score: it.Score,
		}
	}

	res := &openapi.UserActivity{
		TotalUpdates: summary.TotalUpdates,
		Scores:       scores,
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches the scores of a user for a given year
// (GET /users/{userId}/scores/{year})
func (s *Server) ProfileYearlyScoresByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.profileQueryService.YearlyScoresForUser(ctx.Request().Context(), &profilequery.YearlyScoresForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Logger().Errorf("could not fetch scores: %w", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.Score, len(summary.Scores))
	for i, it := range summary.Scores {
		scores[i] = openapi.Score{
			LanguageCode: it.LanguageCode,
			Score:        it.Score,
			LanguageName: it.LanguageName,
		}
	}

	return ctx.JSON(http.StatusOK, &openapi.ProfileScores{
		OverallScore: summary.OverallScore,
		Scores:       scores,
	})
}

// Fetches the contest registrations of a user for a given year
// (GET /users/{userId}/contest-registrations/{year})
func (s *Server) ProfileYearlyContestRegistrationsByUserID(ctx echo.Context, userId types.UUID, year int) error {
	regs, err := s.contestQueryService.YearlyContestRegistrations(ctx.Request().Context(), &contestquery.YearlyContestRegistrationsRequest{
		UserID: userId,
		Year:   year,
	})
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

	for i, it := range regs.Registrations {
		it := it
		res.Registrations[i] = *contestRegistrationToAPI(&it)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches a activity split summary of a user for a given year
// (GET /users/{userId}/activity-split/{year})
func (s *Server) ProfileYearlyActivitySplitByUserID(ctx echo.Context, userId types.UUID, year int) error {
	summary, err := s.profileQueryService.YearlyActivitySplitForUser(ctx.Request().Context(), &profilequery.YearlyActivitySplitForUserRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	scores := make([]openapi.ActivitySplitScore, len(summary.Activities))
	for i, it := range summary.Activities {
		scores[i] = openapi.ActivitySplitScore{
			ActivityId:   it.ActivityID,
			ActivityName: it.ActivityName,
			Score:        it.Score,
		}
	}

	res := &openapi.ActivitySplit{
		Activities: scores,
	}

	return ctx.JSON(http.StatusOK, res)
}

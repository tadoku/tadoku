package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the contest registrations of a user for a given year
// (GET /users/{userId}/contest-registrations/{year})
func (s *Server) ProfileYearlyContestRegistrationsByUserID(ctx echo.Context, userId types.UUID, year int) error {
	regs, err := s.registrationListYearly.Execute(ctx.Request().Context(), &domain.RegistrationListYearlyRequest{
		UserID: userId,
		Year:   year,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
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
		res.Registrations[i] = *contestRegistrationToAPI(&it)
	}

	return ctx.JSON(http.StatusOK, res)
}

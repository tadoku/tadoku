package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches all the ongoing contest registrations of the logged in user, always in a single page
// (GET /contests/configuration-options)
func (s *Server) ContestFindOngoingRegistrations(ctx echo.Context) error {
	regs, err := s.registrationListOngoing.Execute(ctx.Request().Context())
	if err != nil {
		if handled, respErr := handleCommonDomainError(ctx, err); handled {
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

	for i, r := range regs.Registrations {
		r := r
		res.Registrations[i] = *contestRegistrationToAPI(&r)
	}

	return ctx.JSON(http.StatusOK, res)
}

package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Previews platform and contest scores without creating a log
// (POST /logs/score-preview)
func (s *Server) ScorePreview(ctx echo.Context) error {
	var req openapi.ScorePreviewJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var registrationIDs []uuid.UUID
	if req.RegistrationIds != nil {
		registrationIDs = *req.RegistrationIds
	}
	result, err := s.scorePreview.Execute(ctx.Request().Context(), &domain.ScorePreviewRequest{
		RegistrationIDs: registrationIDs,
		UnitID:          req.UnitId,
		UnitKey:         req.UnitKey,
		ActivityID:      req.ActivityId,
		LanguageCode:    req.LanguageCode,
		Amount:          req.Amount,
		DurationSeconds: req.DurationSeconds,
		Tags:            req.Tags,
	})
	if err != nil {
		if handled, responseErr := handleCommonErrors(ctx, err); handled {
			return responseErr
		}
		if errors.Is(err, domain.ErrInvalidLog) {
			return ctx.NoContent(http.StatusBadRequest)
		}
		ctx.Echo().Logger.Error("could not preview score: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	response := openapi.ScorePreview{
		Platform: scoreEstimateToAPI(result.Platform),
		Contests: make([]openapi.ContestScoreEstimate, len(result.Contests)),
	}
	for i, contest := range result.Contests {
		response.Contests[i] = openapi.ContestScoreEstimate{
			RegistrationId: contest.RegistrationID,
			ContestId:      contest.ContestID,
			Estimate:       scoreEstimateToAPI(contest.Estimate),
		}
	}
	return ctx.JSON(http.StatusOK, response)
}

func scoreEstimateToAPI(estimate domain.ScoreEstimate) openapi.ScoreEstimate {
	rules := make([]openapi.AppliedScoringRule, len(estimate.Rules))
	for i, rule := range estimate.Rules {
		rules[i] = openapi.AppliedScoringRule{
			RuleId: rule.RuleID,
			Rate:   rule.Rate,
		}
	}
	return openapi.ScoreEstimate{
		Score:     estimate.Score,
		Source:    openapi.ScoreEstimateSource(estimate.Source),
		RuleSetId: estimate.RuleSetID,
		Rules:     rules,
	}
}

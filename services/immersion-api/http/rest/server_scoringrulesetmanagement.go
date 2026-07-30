package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func (s *Server) ScoringRuleSetListPlatform(ctx echo.Context) error {
	ruleSets, err := s.scoringRuleSetManagement.ListPlatform(ctx.Request().Context())
	if err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scoringRuleSetsToAPI(ruleSets))
}

func (s *Server) ScoringRuleSetCreatePlatform(ctx echo.Context) error {
	var body openapi.ScoringRuleSetCreatePlatformJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}
	ruleSet, err := s.scoringRuleSetManagement.CreatePlatformDraft(
		ctx.Request().Context(),
		scoringRuleSetDraftToDomain(body),
	)
	if err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scoringRuleSetToAPI(*ruleSet))
}

func (s *Server) ScoringRuleSetListContest(ctx echo.Context, id uuid.UUID) error {
	ruleSets, err := s.scoringRuleSetManagement.ListContest(ctx.Request().Context(), id)
	if err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scoringRuleSetsToAPI(ruleSets))
}

func (s *Server) ScoringRuleSetCreateContest(ctx echo.Context, id uuid.UUID) error {
	var body openapi.ScoringRuleSetCreateContestJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}
	ruleSet, err := s.scoringRuleSetManagement.CreateContestDraft(
		ctx.Request().Context(),
		id,
		scoringRuleSetDraftToDomain(body),
	)
	if err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scoringRuleSetToAPI(*ruleSet))
}

func (s *Server) ScoringRuleSetPublish(ctx echo.Context, id uuid.UUID) error {
	ruleSet, err := s.scoringRuleSetManagement.Publish(ctx.Request().Context(), id)
	if err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.JSON(http.StatusOK, scoringRuleSetToAPI(*ruleSet))
}

func (s *Server) ScoringRuleSetActivate(ctx echo.Context, id uuid.UUID) error {
	if err := s.scoringRuleSetManagement.Activate(ctx.Request().Context(), id); err != nil {
		return handleScoringRuleSetError(ctx, err)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func handleScoringRuleSetError(ctx echo.Context, err error) error {
	if handled, responseErr := handleCommonErrors(ctx, err); handled {
		return responseErr
	}
	if errors.Is(err, domain.ErrInvalidScoringRuleSet) {
		return ctx.NoContent(http.StatusBadRequest)
	}
	ctx.Echo().Logger.Error("could not manage scoring rule set: ", err)
	return ctx.NoContent(http.StatusInternalServerError)
}

func scoringRuleSetDraftToDomain(body openapi.ScoringRuleSetDraft) *domain.ScoringRuleSetDraftCreateRequest {
	rules := make([]domain.ScoringRule, len(body.Rules))
	for i, rule := range body.Rules {
		rules[i] = domain.ScoringRule{
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityID:   rule.ActivityId,
			UnitKey:      stringValue(rule.UnitKey),
			LanguageCode: stringValue(rule.LanguageCode),
			Tag:          stringValue(rule.Tag),
			ScoreSource:  domain.ScoreSource(rule.ScoreSource),
			Rate:         rule.Rate,
		}
	}
	mode := ""
	if body.Mode != nil {
		mode = string(*body.Mode)
	}
	return &domain.ScoringRuleSetDraftCreateRequest{
		Mode:              domain.ScoringRuleSetMode(mode),
		FallbackRuleSetID: body.FallbackRuleSetId,
		Rules:             rules,
	}
}

func scoringRuleSetsToAPI(ruleSets []domain.ScoringRuleSet) openapi.ScoringRuleSets {
	result := openapi.ScoringRuleSets{RuleSets: make([]openapi.ScoringRuleSet, len(ruleSets))}
	for i, ruleSet := range ruleSets {
		result.RuleSets[i] = scoringRuleSetToAPI(ruleSet)
	}
	return result
}

func scoringRuleSetToAPI(ruleSet domain.ScoringRuleSet) openapi.ScoringRuleSet {
	rules := make([]openapi.ScoringRule, len(ruleSet.Rules))
	for i, rule := range ruleSet.Rules {
		rules[i] = openapi.ScoringRule{
			Id:           &rule.ID,
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityId:   rule.ActivityID,
			UnitKey:      optionalString(rule.UnitKey),
			LanguageCode: optionalString(rule.LanguageCode),
			Tag:          optionalString(rule.Tag),
			ScoreSource:  openapi.ScoringRuleScoreSource(rule.ScoreSource),
			Rate:         rule.Rate,
		}
	}
	var mode *openapi.ScoringRuleSetMode
	if ruleSet.Mode != "" {
		value := openapi.ScoringRuleSetMode(ruleSet.Mode)
		mode = &value
	}
	return openapi.ScoringRuleSet{
		Id:                ruleSet.ID,
		Scope:             openapi.ScoringRuleSetScope(ruleSet.Scope),
		ContestId:         ruleSet.ContestID,
		Version:           ruleSet.Version,
		Status:            openapi.ScoringRuleSetStatus(ruleSet.Status),
		Active:            ruleSet.Active,
		Mode:              mode,
		FallbackRuleSetId: ruleSet.FallbackRuleSetID,
		Rules:             rules,
		CreatedAt:         ruleSet.CreatedAt,
		PublishedAt:       ruleSet.PublishedAt,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

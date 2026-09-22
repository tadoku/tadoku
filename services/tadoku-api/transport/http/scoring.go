package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *server) ImmersionScorePreview(ctx context.Context, request openapi.ImmersionScorePreviewRequestObject) (openapi.ImmersionScorePreviewResponseObject, error) {
	body := request.Body
	parameters := app.ScorePreviewParameters{
		UnitID:          body.UnitId,
		UnitKey:         body.UnitKey,
		ActivityID:      body.ActivityId,
		LanguageCode:    body.LanguageCode,
		Amount:          body.Amount,
		DurationSeconds: body.DurationSeconds,
		Tags:            body.Tags,
	}
	if body.RegistrationIds != nil {
		parameters.RegistrationIDs = *body.RegistrationIds
	}

	result, err := s.application.PreviewScore(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "preview score", err)
		if scoringHTTPError(err) {
			return nil, err
		}
		return openapi.ImmersionScorePreview500Response{}, nil
	}
	return openapi.ImmersionScorePreview200JSONResponse(scorePreviewResponse(result)), nil
}

func (s *server) ImmersionScoringRuleSetListPlatform(ctx context.Context, _ openapi.ImmersionScoringRuleSetListPlatformRequestObject) (openapi.ImmersionScoringRuleSetListPlatformResponseObject, error) {
	sets, err := s.application.ListPlatformScoringRuleSets(ctx)
	if err != nil {
		s.logOperationError(ctx, "list platform scoring rule sets", err)
		if scoringHTTPError(err) {
			return nil, err
		}
		return openapi.ImmersionScoringRuleSetListPlatform500Response{}, nil
	}
	return openapi.ImmersionScoringRuleSetListPlatform200JSONResponse(scoringRuleSetsResponse(sets)), nil
}

func (s *server) ImmersionScoringRuleSetListContest(ctx context.Context, request openapi.ImmersionScoringRuleSetListContestRequestObject) (openapi.ImmersionScoringRuleSetListContestResponseObject, error) {
	sets, err := s.application.ListContestScoringRuleSets(ctx, request.Id)
	if err != nil {
		s.logOperationError(ctx, "list contest scoring rule sets", err)
		if scoringHTTPError(err) {
			return nil, err
		}
		return openapi.ImmersionScoringRuleSetListContest500Response{}, nil
	}
	return openapi.ImmersionScoringRuleSetListContest200JSONResponse(scoringRuleSetsResponse(sets)), nil
}

func (s *server) ImmersionScoringRuleSetCreatePlatform(ctx context.Context, request openapi.ImmersionScoringRuleSetCreatePlatformRequestObject) (openapi.ImmersionScoringRuleSetCreatePlatformResponseObject, error) {
	created, err := s.application.CreatePlatformScoringRuleSetDraft(ctx, scoringRuleSetDraftParameters(*request.Body))
	if err != nil {
		s.logOperationError(ctx, "create platform scoring rule set", err)
		if scoringHTTPError(err) {
			return nil, err
		}
		return openapi.ImmersionScoringRuleSetCreatePlatform500Response{}, nil
	}
	return openapi.ImmersionScoringRuleSetCreatePlatform200JSONResponse(scoringRuleSetResponse(*created)), nil
}

func (s *server) ImmersionScoringRuleSetCreateContest(ctx context.Context, request openapi.ImmersionScoringRuleSetCreateContestRequestObject) (openapi.ImmersionScoringRuleSetCreateContestResponseObject, error) {
	created, err := s.application.CreateContestScoringRuleSetDraft(ctx, request.Id, scoringRuleSetDraftParameters(*request.Body))
	if err != nil {
		s.logOperationError(ctx, "create contest scoring rule set", err)
		if scoringHTTPError(err) {
			return nil, err
		}
		return openapi.ImmersionScoringRuleSetCreateContest500Response{}, nil
	}
	return openapi.ImmersionScoringRuleSetCreateContest200JSONResponse(scoringRuleSetResponse(*created)), nil
}

func scoringRuleSetDraftParameters(body openapi.ImmersionScoringRuleSetDraft) app.ScoringRuleSetDraftParameters {
	parameters := app.ScoringRuleSetDraftParameters{
		FallbackRuleSetID: body.FallbackRuleSetId,
		Rules:             make([]app.ScoringRule, len(body.Rules)),
	}
	if body.Mode != nil {
		parameters.Mode = string(*body.Mode)
	}
	for i, rule := range body.Rules {
		parameters.Rules[i] = app.ScoringRule{
			Priority:     rule.Priority,
			Stackable:    rule.Stackable,
			ActivityID:   rule.ActivityId,
			UnitKey:      scoringString(rule.UnitKey),
			LanguageCode: scoringString(rule.LanguageCode),
			Tag:          scoringString(rule.Tag),
			Source:       app.ScoringSource(rule.ScoreSource),
			Rate:         rule.Rate,
		}
	}
	return parameters
}

func scorePreviewResponse(result *app.ScorePreview) openapi.ImmersionScorePreview {
	response := openapi.ImmersionScorePreview{
		Platform: scoreEstimateResponse(result.Platform),
		Contests: make([]openapi.ImmersionContestScoreEstimate, len(result.Contests)),
	}
	for i, contest := range result.Contests {
		response.Contests[i] = openapi.ImmersionContestScoreEstimate{
			RegistrationId: contest.RegistrationID,
			ContestId:      contest.ContestID,
			Estimate:       scoreEstimateResponse(contest.Estimate),
		}
	}
	return response
}

func scoreEstimateResponse(estimate app.ScoreEstimate) openapi.ImmersionScoreEstimate {
	rules := make([]openapi.ImmersionAppliedScoringRule, len(estimate.Rules))
	for i, rule := range estimate.Rules {
		rules[i] = openapi.ImmersionAppliedScoringRule{
			RuleId: rule.RuleID,
			Rate:   rule.Rate,
		}
	}
	return openapi.ImmersionScoreEstimate{
		Score:     estimate.Score,
		Source:    openapi.ImmersionScoreEstimateSource(estimate.Source),
		RuleSetId: estimate.RuleSetID,
		Rules:     rules,
	}
}

func scoringHTTPError(err error) bool {
	kind := errx.KindOf(err)
	return kind != errx.Unknown && kind != errx.Internal
}

func scoringRuleSetsResponse(sets []app.ScoringRuleSet) openapi.ImmersionScoringRuleSets {
	response := openapi.ImmersionScoringRuleSets{
		RuleSets: make([]openapi.ImmersionScoringRuleSet, len(sets)),
	}
	for i, set := range sets {
		rules := make([]openapi.ImmersionScoringRule, len(set.Rules))
		for j, rule := range set.Rules {
			rules[j] = openapi.ImmersionScoringRule{
				Id:           &rule.ID,
				Priority:     rule.Priority,
				Stackable:    rule.Stackable,
				ActivityId:   rule.ActivityID,
				UnitKey:      optionalScoringString(rule.UnitKey),
				LanguageCode: optionalScoringString(rule.LanguageCode),
				Tag:          optionalScoringString(rule.Tag),
				ScoreSource:  openapi.ImmersionScoringRuleScoreSource(rule.Source),
				Rate:         rule.Rate,
			}
		}
		var mode *openapi.ImmersionScoringRuleSetMode
		if set.Mode != "" {
			value := openapi.ImmersionScoringRuleSetMode(set.Mode)
			mode = &value
		}
		response.RuleSets[i] = openapi.ImmersionScoringRuleSet{
			Id:                set.ID,
			Scope:             openapi.ImmersionScoringRuleSetScope(set.Scope),
			ContestId:         set.ContestID,
			Version:           set.Version,
			Status:            openapi.ImmersionScoringRuleSetStatus(set.Status),
			Active:            set.Active,
			Mode:              mode,
			FallbackRuleSetId: set.FallbackRuleSetID,
			Rules:             rules,
			CreatedAt:         set.CreatedAt,
			PublishedAt:       set.PublishedAt,
		}
	}
	return response
}

func scoringRuleSetResponse(set app.ScoringRuleSet) openapi.ImmersionScoringRuleSet {
	return scoringRuleSetsResponse([]app.ScoringRuleSet{set}).RuleSets[0]
}

func optionalScoringString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func scoringString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

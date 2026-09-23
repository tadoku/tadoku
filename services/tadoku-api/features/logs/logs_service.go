package logs

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
)

type Service struct {
	logs                 *LogsRepository
	scoringEngineEnabled bool
}

func NewService(logs *LogsRepository, scoringEngineEnabled bool) *Service {
	return &Service{logs: logs, scoringEngineEnabled: scoringEngineEnabled}
}

func (s *Service) ScoringEngineEnabled() bool { return s.scoringEngineEnabled }

func (s *Service) ConfigurationOptions(ctx context.Context, userID uuid.UUID) (*ConfigurationOptions, error) {
	units, err := s.logs.ListUnits(ctx)
	if err != nil {
		return nil, err
	}

	codes, err := s.logs.ListUserLanguageCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ConfigurationOptions{
		Units:                units,
		UserLanguageCodes:    codes,
		ScoringEngineEnabled: s.scoringEngineEnabled,
	}, nil
}

var defaultTagSuggestions = []string{
	"anime",
	"audiobook",
	"book",
	"chat",
	"chorusing",
	"comic",
	"conversation",
	"drama",
	"ebook",
	"fiction",
	"game",
	"grammar",
	"lyric",
	"news",
	"non-fiction",
	"online video",
	"podcast",
	"presentation",
	"shadowing",
	"social media",
	"srs",
	"textbook",
	"tv",
	"vocabulary",
	"web page",
	"youtube",
	"nsfw",
}

func (s *Service) TagSuggestions(ctx context.Context, userID uuid.UUID, query string) ([]TagSuggestion, error) {
	suggestions, err := s.logs.ListTagSuggestions(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	// Preserve the legacy append limit: thirty history results may gain one default.
	seen := make(map[string]struct{})
	for _, s := range suggestions {
		seen[strings.ToLower(s.Tag)] = struct{}{}
	}
	for _, tag := range matchingDefaultTags(query) {
		if _, exists := seen[strings.ToLower(tag)]; !exists {
			seen[strings.ToLower(tag)] = struct{}{}
			suggestions = append(suggestions, TagSuggestion{Tag: tag, Count: 0})
			if len(suggestions) >= 30 {
				break
			}
		}
	}

	// Keep NSFW below ordinary suggestions, including frequently used tags.
	sort.SliceStable(suggestions, func(i, j int) bool {
		return !strings.EqualFold(suggestions[i].Tag, "nsfw") && strings.EqualFold(suggestions[j].Tag, "nsfw")
	})

	return suggestions, nil
}

func matchingDefaultTags(query string) []string {
	query = strings.ToLower(query)
	if query == "" {
		return defaultTagSuggestions
	}

	tags := []string{}
	for _, tag := range defaultTagSuggestions {
		if strings.Contains(tag, query) {
			tags = append(tags, tag)
		}
	}
	return tags
}

func (s *Service) YearlyActivity(ctx context.Context, userID uuid.UUID, year int) (*YearlyActivity, error) {
	scores, err := s.logs.YearlyActivity(ctx, userID, int16(year))
	if err != nil {
		return nil, err
	}

	result := &YearlyActivity{Scores: scores}
	for _, score := range scores {
		result.TotalUpdates += score.Updates
	}

	return result, nil
}

func (s *Service) YearlyScores(ctx context.Context, userID uuid.UUID, year int) (*YearlyScores, error) {
	scores, err := s.logs.YearlyScores(ctx, userID, int16(year))
	if err != nil {
		return nil, err
	}

	result := &YearlyScores{Scores: scores}
	for _, score := range scores {
		result.OverallScore += score.Score
	}

	return result, nil
}

func (s *Service) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int) ([]ActivitySplitScore, error) {
	scores, err := s.logs.YearlyActivitySplit(ctx, userID, int16(year))
	if err != nil {
		return nil, err
	}
	for i := range scores {
		found := false
		for _, activity := range activities.All() {
			if int(activity.ID) == scores[i].ActivityID {
				scores[i].ActivityName = activity.Name
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid activity %d", scores[i].ActivityID)
		}
	}
	return scores, nil
}

func (s *Service) ContestScores(ctx context.Context, userID, contestID uuid.UUID) ([]Score, error) {
	return s.logs.ContestScores(ctx, userID, contestID)
}

func (s *Service) ContestActivity(ctx context.Context, userID, contestID uuid.UUID) ([]ContestActivity, error) {
	return s.logs.ContestActivity(ctx, userID, contestID)
}

func (s *Service) FindLog(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Log, error) {
	log, err := s.logs.FindLog(ctx, id, includeDeleted)
	if err != nil {
		return nil, err
	}

	log.Registrations, err = s.logs.AttachedRegistrations(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := hydrateLogActivity(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *Service) ResolveTracking(ctx context.Context, activityID int32, languageCode string, unitID *uuid.UUID, unitKey *string, amount *float32, duration *int32) (Tracking, error) {
	if activityID < 1 || activityID > 5 {
		return Tracking{}, ErrInvalidActivity
	}
	var unit *Unit
	var err error
	if amount != nil && unitID != nil {
		unit, err = s.logs.FindTrackingUnit(ctx, unitID, nil, activityID, languageCode)
	} else if amount != nil && unitKey != nil {
		unit, err = s.logs.FindTrackingUnit(ctx, nil, unitKey, activityID, languageCode)
	}
	if err != nil {
		return Tracking{}, err
	}
	return ValidateAndResolveTracking(activityID, unit, unitID, unitKey, amount, duration)
}

func (s *Service) Create(ctx context.Context, mutation Mutation) error {
	if err := s.logs.CreateLog(ctx, mutation); err != nil {
		return err
	}
	for _, tracking := range mutation.ContestTrackings {
		if err := s.logs.CreateContestLog(ctx, mutation.ID, tracking); err != nil {
			return err
		}
	}
	for _, tag := range mutation.Tags {
		if err := s.logs.InsertTag(ctx, mutation.ID, mutation.UserID, tag); err != nil {
			return err
		}
	}

	seen := make(map[uuid.UUID]struct{}, len(mutation.ContestTrackings))
	for _, tracking := range mutation.ContestTrackings {
		if _, exists := seen[tracking.ContestID]; exists {
			continue
		}
		seen[tracking.ContestID] = struct{}{}
		contestID := tracking.ContestID
		if err := s.logs.InsertOutbox(ctx, mutation.UserID, &contestID, nil, "refresh_contest_score"); err != nil {
			return err
		}
	}
	if mutation.EligibleOfficialLeaderboard {
		year := mutation.Year
		return s.logs.InsertOutbox(ctx, mutation.UserID, nil, &year, "refresh_official_scores")
	}
	return nil
}
func (s *Service) Update(ctx context.Context, mutation Mutation) error {
	if err := s.logs.LockLog(ctx, mutation.ID); err != nil {
		return err
	}
	outbox, err := s.logs.OutboxContext(ctx, mutation.ID)
	if err != nil {
		return err
	}
	if err := s.logs.UpdateLog(ctx, mutation); err != nil {
		return err
	}
	if len(mutation.ContestTrackings) == 0 {
		inherited := mutation.Tracking
		inherited.RuleSetID = nil
		inherited.RuleIDs = nil
		inherited.Rates = nil
		inherited.Source = ""
		if err := s.logs.UpdateOngoingContestLogs(ctx, mutation.ID, inherited, mutation.Now); err != nil {
			return err
		}
	} else {
		for _, tracking := range mutation.ContestTrackings {
			if err := s.logs.UpdateContestLog(ctx, mutation.ID, tracking, mutation.Now); err != nil {
				return err
			}
		}
	}
	if err := s.logs.DeleteTags(ctx, mutation.ID); err != nil {
		return err
	}
	for _, tag := range mutation.Tags {
		if err := s.logs.InsertTag(ctx, mutation.ID, outbox.UserID, tag); err != nil {
			return err
		}
	}
	contestIDs, err := s.logs.OngoingContestIDs(ctx, mutation.ID, mutation.Now)
	if err != nil {
		return err
	}
	for _, id := range contestIDs {
		contestID := id
		if err := s.logs.InsertOutbox(ctx, outbox.UserID, &contestID, nil, "refresh_contest_score"); err != nil {
			return err
		}
	}
	if outbox.EligibleOfficial {
		year := outbox.Year
		return s.logs.InsertOutbox(ctx, outbox.UserID, nil, &year, "refresh_official_scores")
	}
	return nil
}

func (s *Service) RegistrationsForRescoring(log *Log, now time.Time) []RegistrationReference {
	if !s.scoringEngineEnabled {
		return nil
	}
	selected := make([]RegistrationReference, 0, len(log.Registrations))
	for _, registration := range log.Registrations {
		if !registration.ContestEnd.Before(now) {
			selected = append(selected, registration)
		}
	}
	return selected
}

func (s *Service) ListUserLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	result, err := s.logs.ListUserLogs(ctx, parameters.normalized())
	if err != nil {
		return nil, err
	}
	for i := range result.Logs {
		if err := hydrateLogActivity(&result.Logs[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Service) ListContestLogs(ctx context.Context, parameters ListParameters) (*LogList, error) {
	result, err := s.logs.ListContestLogs(ctx, parameters.normalized())
	if err != nil {
		return nil, err
	}
	for i := range result.Logs {
		if err := hydrateLogActivity(&result.Logs[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

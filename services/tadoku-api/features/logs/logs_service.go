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

func (s *Service) LockLog(ctx context.Context, id uuid.UUID) error { return s.logs.LockLog(ctx, id) }
func (s *Service) Create(ctx context.Context, mutation Mutation) error {
	return s.logs.CreateLog(ctx, mutation)
}
func (s *Service) CreateContest(ctx context.Context, logID uuid.UUID, tracking ContestTracking) error {
	return s.logs.CreateContestLog(ctx, logID, tracking)
}
func (s *Service) InsertTag(ctx context.Context, logID, userID uuid.UUID, tag string) error {
	return s.logs.InsertTag(ctx, logID, userID, tag)
}
func (s *Service) DeleteTags(ctx context.Context, logID uuid.UUID) error {
	return s.logs.DeleteTags(ctx, logID)
}
func (s *Service) InsertOutbox(ctx context.Context, userID uuid.UUID, contestID *uuid.UUID, year *int16, event string) error {
	return s.logs.InsertOutbox(ctx, userID, contestID, year, event)
}
func (s *Service) OutboxContext(ctx context.Context, id uuid.UUID) (OutboxContext, error) {
	return s.logs.OutboxContext(ctx, id)
}
func (s *Service) Update(ctx context.Context, mutation Mutation) error {
	return s.logs.UpdateLog(ctx, mutation)
}
func (s *Service) UpdateContest(ctx context.Context, logID uuid.UUID, tracking ContestTracking, now time.Time) error {
	return s.logs.UpdateContestLog(ctx, logID, tracking, now)
}
func (s *Service) UpdateOngoingContests(ctx context.Context, logID uuid.UUID, tracking Tracking, now time.Time) error {
	return s.logs.UpdateOngoingContestLogs(ctx, logID, tracking, now)
}
func (s *Service) OngoingContestIDs(ctx context.Context, id uuid.UUID, now time.Time) ([]uuid.UUID, error) {
	return s.logs.OngoingContestIDs(ctx, id, now)
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

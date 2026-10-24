package logs

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	logs                 *LogsRepository
	scoringEngineEnabled bool
}

func NewService(logs *LogsRepository, scoringEngineEnabled bool) *Service {
	return &Service{logs: logs, scoringEngineEnabled: scoringEngineEnabled}
}

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
	if err := validateYear(year); err != nil {
		return nil, err
	}

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
	if err := validateYear(year); err != nil {
		return nil, err
	}

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
	if err := validateYear(year); err != nil {
		return nil, err
	}

	scores, err := s.logs.YearlyActivitySplit(ctx, userID, int16(year))
	if err != nil {
		return nil, err
	}
	for i := range scores {
		activity, err := findActivity(int32(scores[i].ActivityID))
		if err != nil {
			return nil, err
		}
		scores[i].ActivityName = activity.Name
	}
	return scores, nil
}

func (s *Service) ContestScores(ctx context.Context, userID, contestID uuid.UUID) ([]Score, error) {
	return s.logs.ContestScores(ctx, userID, contestID)
}

func (s *Service) ContestActivity(ctx context.Context, userID, contestID uuid.UUID) ([]ContestActivity, error) {
	return s.logs.ContestActivity(ctx, userID, contestID)
}

// FindLog applies no visibility rules; callers must authorize the actor.
func (s *Service) FindLog(ctx context.Context, id uuid.UUID) (*Log, error) {
	log, err := s.logs.FindLog(ctx, id, false)
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

func (s *Service) FindLogForViewer(ctx context.Context, id uuid.UUID, parameters FindForViewerParameters) (*Log, error) {
	log, err := s.logs.FindLog(ctx, id, parameters.IncludeDeleted)
	if err != nil {
		return nil, err
	}

	if mayViewRegistrations(parameters.Viewer, log.UserID) {
		log.Registrations, err = s.logs.AttachedRegistrations(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	if err := hydrateLogActivity(log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *Service) ListUserLogs(ctx context.Context, parameters ListParameters, createdOn *time.Time) (*LogList, error) {
	if err := parameters.Validate(); err != nil {
		return nil, err
	}

	result, err := s.logs.ListUserLogs(ctx, parameters.normalized(), createdOn)
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
	if err := parameters.Validate(); err != nil {
		return nil, err
	}

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

const exportBatchSize = 500

func (s *Service) ExportUserLogs(ctx context.Context, userID uuid.UUID) ([]ExportedLog, error) {
	result := make([]ExportedLog, 0)
	var after *exportCursor
	for {
		batch, err := s.logs.ListActiveUserLogsAfter(ctx, userID, after, exportBatchSize)
		if err != nil {
			return nil, err
		}

		for i := range batch {
			batch[i].Activity, err = findActivity(batch[i].Activity.ID)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, batch...)
		if len(batch) < exportBatchSize {
			return result, nil
		}

		last := batch[len(batch)-1]
		after = &exportCursor{createdAt: last.CreatedAt, id: last.ID}
	}
}

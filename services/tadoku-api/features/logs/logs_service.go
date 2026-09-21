package logs

import (
	"context"
	"strings"

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

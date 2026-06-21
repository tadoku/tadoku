package domain

import (
	"context"
	"strings"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type TagSuggestion struct {
	Tag   string
	Count int
}

type TagSuggestionsRepository interface {
	FetchTagSuggestionsForUser(ctx context.Context, userID uuid.UUID, query string) ([]TagSuggestion, error)
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
}

type TagSuggestionsRequest struct {
	Query string
}

type TagSuggestionsResponse struct {
	Suggestions []TagSuggestion
}

type TagSuggestions struct {
	repo TagSuggestionsRepository
}

func NewTagSuggestions(repo TagSuggestionsRepository) *TagSuggestions {
	return &TagSuggestions{
		repo: repo,
	}
}

func (s *TagSuggestions) Execute(ctx context.Context, req *TagSuggestionsRequest) (*TagSuggestionsResponse, error) {
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil || session.Subject == "" {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(session.Subject)
	if err != nil {
		return nil, ErrUnauthorized
	}

	suggestions, err := s.repo.FetchTagSuggestionsForUser(ctx, userID, req.Query)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	for _, s := range suggestions {
		seen[strings.ToLower(s.Tag)] = struct{}{}
	}
	for _, tag := range matchingDefaultTags(req.Query) {
		if _, exists := seen[strings.ToLower(tag)]; !exists {
			seen[strings.ToLower(tag)] = struct{}{}
			suggestions = append(suggestions, TagSuggestion{Tag: tag, Count: 0})
			if len(suggestions) >= 30 {
				break
			}
		}
	}

	return &TagSuggestionsResponse{
		Suggestions: suggestions,
	}, nil
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

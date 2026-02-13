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
	FetchDefaultTagsMatching(ctx context.Context, query string) ([]string, error)
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
	// Get user ID from session if authenticated
	var userID *uuid.UUID
	session := commondomain.ParseUserIdentity(ctx)
	if session != nil && session.Subject != "" {
		id, err := uuid.Parse(session.Subject)
		if err == nil {
			userID = &id
		}
	}

	var suggestions []TagSuggestion

	// If user is authenticated, fetch their previously used tags
	if userID != nil {
		userTags, err := s.repo.FetchTagSuggestionsForUser(ctx, *userID, req.Query)
		if err != nil {
			return nil, err
		}
		suggestions = userTags
	}

	// Append default tags, deduplicating against user tags
	defaultTags, err := s.repo.FetchDefaultTagsMatching(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	for _, s := range suggestions {
		seen[strings.ToLower(s.Tag)] = struct{}{}
	}
	for _, tag := range defaultTags {
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

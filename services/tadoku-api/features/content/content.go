// Package content owns editorial content and its persistence.
package content

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Announcement struct {
	ID                                     uuid.UUID
	Namespace, Title, Content, Style       string
	Href                                   *string
	StartsAt, EndsAt, CreatedAt, UpdatedAt time.Time
}

var ErrInvalidNamespace = errors.New("namespace is required")

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) ActiveAnnouncements(ctx context.Context, namespace string) ([]Announcement, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	// Publication policy belongs here; persistence only applies these inputs.
	return s.repository.ListActiveAnnouncements(ctx, namespace, timex.Now(), 10)
}

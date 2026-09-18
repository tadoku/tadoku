package announcements

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (s *Service) DeleteAnnouncement(ctx context.Context, namespace string, id uuid.UUID) error {
	return s.announcements.DeleteAnnouncement(ctx, namespace, id, timex.Now())
}

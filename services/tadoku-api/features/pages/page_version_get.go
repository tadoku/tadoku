package pages

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetPageVersion(ctx context.Context, namespace string, pageID, contentID uuid.UUID) (*PageVersion, error) {
	return s.pages.GetPageVersion(ctx, namespace, pageID, contentID)
}

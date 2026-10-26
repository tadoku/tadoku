package contests

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID, now time.Time) (RemovalResult, error) {
	return RemovalResult{}, nil
}

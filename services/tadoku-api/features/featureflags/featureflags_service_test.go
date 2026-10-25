package featureflags

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestGetNamedUserAccessMapsClientErrorToUnavailable(t *testing.T) {
	t.Parallel()

	service := NewService(nil, fliptmanagement.NewClient(fliptmanagement.Config{}))
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")

	_, err := service.GetNamedUserAccess(t.Context(), releaseLogEntryV2, userID)
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("GetNamedUserAccess() error kind = %v, want unavailable", errx.KindOf(err))
	}
}

func TestGrantMapsClientErrorToUnavailable(t *testing.T) {
	t.Parallel()

	service := NewService(nil, fliptmanagement.NewClient(fliptmanagement.Config{}))
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")

	_, err := service.Grant(t.Context(), releaseLogEntryV2, userID)
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("Grant() error kind = %v, want unavailable", errx.KindOf(err))
	}
}

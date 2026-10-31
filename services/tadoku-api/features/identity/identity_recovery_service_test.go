package identity

import (
	"context"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestRecoveryWithoutTenant(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, "tadoku.app", nil)
	ctx := context.Background()

	_, admitErr := service.AdmitRecovery(ctx, Recovery{Email: "user@example.com", IP: "192.0.2.1"})
	_, settingsErr := service.RecoverySettingsURL(ctx, nil)

	if errx.KindOf(admitErr) != errx.Unavailable {
		t.Errorf("AdmitRecovery error = %v, want unavailable", admitErr)
	}
	if errx.KindOf(settingsErr) != errx.Unavailable {
		t.Errorf("RecoverySettingsURL error = %v, want unavailable", settingsErr)
	}
}

package content_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

func TestEmptyNamespaceIsRejectedBeforeStorage(t *testing.T) {
	service := content.NewService(nil)
	_, err := service.ListActiveAnnouncements(context.Background(), "")
	if !errors.Is(err, content.ErrInvalidNamespace) {
		t.Errorf("error=%v want invalid namespace", err)
	}
}

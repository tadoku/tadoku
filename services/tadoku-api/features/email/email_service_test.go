package email

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestFindRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, nil).Find(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("Find without tenant = %v, want unavailable", err)
	}
}

func TestWriteRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, nil).Write(t.Context(), mail.Message{})
	if errx.KindOf(err) != errx.Unavailable {
		t.Errorf("Write without tenant = %v, want unavailable", err)
	}
}

package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockAnnouncementDeleteRepo struct {
	deleteAnnouncementFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAnnouncementDeleteRepo) DeleteAnnouncement(ctx context.Context, id uuid.UUID) error {
	if m.deleteAnnouncementFn != nil {
		return m.deleteAnnouncementFn(ctx, id)
	}
	return nil
}

func TestAnnouncementDelete_Execute(t *testing.T) {
	t.Run("deletes announcement successfully", func(t *testing.T) {
		id := uuid.New()
		var deletedID uuid.UUID
		repo := &mockAnnouncementDeleteRepo{
			deleteAnnouncementFn: func(ctx context.Context, reqID uuid.UUID) error {
				deletedID = reqID
				return nil
			},
		}

		svc := contentdomain.NewAnnouncementDelete(repo)
		err := svc.Execute(adminContext(), id)

		assert.NoError(t, err)
		assert.Equal(t, id, deletedID)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementDeleteRepo{}
		svc := contentdomain.NewAnnouncementDelete(repo)

		err := svc.Execute(userContext(), uuid.New())

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})
}

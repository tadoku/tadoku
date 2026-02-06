package domain_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockAnnouncementDeleteRepo struct {
	deleteAnnouncementFn func(ctx context.Context, id uuid.UUID, namespace string) error
}

func (m *mockAnnouncementDeleteRepo) DeleteAnnouncement(ctx context.Context, id uuid.UUID, namespace string) error {
	if m.deleteAnnouncementFn != nil {
		return m.deleteAnnouncementFn(ctx, id, namespace)
	}
	return nil
}

func TestAnnouncementDelete_Execute(t *testing.T) {
	t.Run("deletes announcement successfully", func(t *testing.T) {
		id := uuid.New()
		var deletedID uuid.UUID
		var deletedNamespace string
		repo := &mockAnnouncementDeleteRepo{
			deleteAnnouncementFn: func(ctx context.Context, reqID uuid.UUID, namespace string) error {
				deletedID = reqID
				deletedNamespace = namespace
				return nil
			},
		}

		svc := contentdomain.NewAnnouncementDelete(repo)
		err := svc.Execute(adminContext(), id, "tadoku")

		assert.NoError(t, err)
		assert.Equal(t, id, deletedID)
		assert.Equal(t, "tadoku", deletedNamespace)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementDeleteRepo{}
		svc := contentdomain.NewAnnouncementDelete(repo)

		err := svc.Execute(userContext(), uuid.New(), "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})
}

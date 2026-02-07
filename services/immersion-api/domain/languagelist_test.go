package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLanguageListRepo struct {
	languages []domain.Language
	listErr   error
}

func (m *mockLanguageListRepo) ListLanguages(_ context.Context) ([]domain.Language, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.languages, nil
}

func TestLanguageList(t *testing.T) {
	adminCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000001",
		Role:    commondomain.RoleAdmin,
	})
	userCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000002",
		Role:    commondomain.RoleUser,
	})
	guestCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000003",
		Role:    commondomain.RoleGuest,
	})

	t.Run("admin can list languages", func(t *testing.T) {
		repo := &mockLanguageListRepo{languages: []domain.Language{
			{Code: "jpn", Name: "Japanese"},
			{Code: "eng", Name: "English"},
		}}
		svc := domain.NewLanguageList(repo)

		result, err := svc.Execute(adminCtx)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "jpn", result[0].Code)
		assert.Equal(t, "eng", result[1].Code)
	})

	t.Run("admin gets empty list when no languages exist", func(t *testing.T) {
		repo := &mockLanguageListRepo{languages: []domain.Language{}}
		svc := domain.NewLanguageList(repo)

		result, err := svc.Execute(adminCtx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("non-admin cannot list languages", func(t *testing.T) {
		repo := &mockLanguageListRepo{}
		svc := domain.NewLanguageList(repo)

		_, err := svc.Execute(userCtx)
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("guest cannot list languages", func(t *testing.T) {
		repo := &mockLanguageListRepo{}
		svc := domain.NewLanguageList(repo)

		_, err := svc.Execute(guestCtx)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})
}

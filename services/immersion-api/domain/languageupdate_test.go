package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLanguageUpdateRepo struct {
	languages map[string]string
	updateErr error
}

func (m *mockLanguageUpdateRepo) UpdateLanguage(_ context.Context, code string, name string) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.languages[code] = name
	return nil
}

func (m *mockLanguageUpdateRepo) LanguageExists(_ context.Context, code string) (bool, error) {
	_, ok := m.languages[code]
	return ok, nil
}

func TestLanguageUpdate(t *testing.T) {
	adminCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000001",
		Role:    commondomain.RoleAdmin,
	})
	userCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000002",
		Role:    commondomain.RoleUser,
	})

	t.Run("admin can update a language", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageUpdateRequest{
			Code: "jpn",
			Name: "Japanese (updated)",
		})
		require.NoError(t, err)
		assert.Equal(t, "Japanese (updated)", repo.languages["jpn"])
	})

	t.Run("non-admin cannot update a language", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(userCtx, &domain.LanguageUpdateRequest{
			Code: "jpn",
			Name: "Japanese (updated)",
		})
		assert.Error(t, err)
	})

	t.Run("updating non-existent language returns not found", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageUpdateRequest{
			Code: "xxx",
			Name: "Unknown",
		})
		assert.Error(t, err)
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageUpdateRequest{
			Code: "jpn",
			Name: "",
		})
		assert.Error(t, err)
	})
}

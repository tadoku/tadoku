package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLanguageCreateRepo struct {
	languages map[string]string
	createErr error
}

func (m *mockLanguageCreateRepo) CreateLanguage(_ context.Context, code string, name string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.languages[code] = name
	return nil
}

func (m *mockLanguageCreateRepo) LanguageExists(_ context.Context, code string) (bool, error) {
	_, ok := m.languages[code]
	return ok, nil
}

func TestLanguageCreate(t *testing.T) {
	adminCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000001",
		Role:    commondomain.RoleAdmin,
	})
	userCtx := ctxWithToken(&commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000002",
		Role:    commondomain.RoleUser,
	})

	t.Run("admin can create a language", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "Japanese",
		})
		require.NoError(t, err)
		assert.Equal(t, "Japanese", repo.languages["jpn"])
	})

	t.Run("non-admin cannot create a language", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(userCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "Japanese",
		})
		assert.Error(t, err)
	})

	t.Run("duplicate code returns conflict", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "Japanese",
		})
		assert.Error(t, err)
	})

	t.Run("empty code returns validation error", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "",
			Name: "Japanese",
		})
		assert.Error(t, err)
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "",
		})
		assert.Error(t, err)
	})
}

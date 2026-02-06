package domain_test

import (
	"context"
	"testing"

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
	adminCtx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000001",
		Role:    commondomain.RoleAdmin,
	})
	userCtx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{
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
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if repo.languages["jpn"] != "Japanese" {
			t.Fatal("expected language to be created")
		}
	})

	t.Run("non-admin cannot create a language", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(userCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "Japanese",
		})
		if err == nil {
			t.Fatal("expected error for non-admin")
		}
	})

	t.Run("duplicate code returns conflict", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "Japanese",
		})
		if err == nil {
			t.Fatal("expected conflict error")
		}
	})

	t.Run("empty code returns validation error", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "",
			Name: "Japanese",
		})
		if err == nil {
			t.Fatal("expected validation error for empty code")
		}
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		repo := &mockLanguageCreateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageCreate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageCreateRequest{
			Code: "jpn",
			Name: "",
		})
		if err == nil {
			t.Fatal("expected validation error for empty name")
		}
	})
}

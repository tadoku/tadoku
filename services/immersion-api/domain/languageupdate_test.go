package domain_test

import (
	"context"
	"testing"

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
	adminCtx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{
		Subject: "00000000-0000-0000-0000-000000000001",
		Role:    commondomain.RoleAdmin,
	})
	userCtx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{
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
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if repo.languages["jpn"] != "Japanese (updated)" {
			t.Fatal("expected language to be updated")
		}
	})

	t.Run("non-admin cannot update a language", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(userCtx, &domain.LanguageUpdateRequest{
			Code: "jpn",
			Name: "Japanese (updated)",
		})
		if err == nil {
			t.Fatal("expected error for non-admin")
		}
	})

	t.Run("updating non-existent language returns not found", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageUpdateRequest{
			Code: "xxx",
			Name: "Unknown",
		})
		if err == nil {
			t.Fatal("expected not found error")
		}
	})

	t.Run("empty name returns validation error", func(t *testing.T) {
		repo := &mockLanguageUpdateRepo{languages: map[string]string{"jpn": "Japanese"}}
		svc := domain.NewLanguageUpdate(repo)

		err := svc.Execute(adminCtx, &domain.LanguageUpdateRequest{
			Code: "jpn",
			Name: "",
		})
		if err == nil {
			t.Fatal("expected validation error for empty name")
		}
	})
}

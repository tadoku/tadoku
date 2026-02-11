package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type mockAuthorizationReader struct {
	allowed bool
	err     error
	called  bool
	subject ketoclient.Subject
}

func (m *mockAuthorizationReader) CheckPermission(ctx context.Context, namespace, object, relation string, subject ketoclient.Subject) (bool, error) {
	m.called = true
	m.subject = subject
	return m.allowed, m.err
}

func (m *mockAuthorizationReader) CheckPermissions(ctx context.Context, checks []ketoclient.PermissionCheck) []ketoclient.PermissionResult {
	return nil
}

func (m *mockAuthorizationReader) ListSubjectIDsForRelation(ctx context.Context, namespace, object, relation string) ([]string, error) {
	return nil, nil
}

func TestPublicPermissionCheck_Execute(t *testing.T) {
	allowlist, err := domain.ParsePermissionAllowlist("app:view")
	require.NoError(t, err)

	authCtx := commonroles.WithClaims(context.Background(), commonroles.Claims{
		Subject:       "user-1",
		Authenticated: true,
	})

	t.Run("requires authenticated subject", func(t *testing.T) {
		keto := &mockAuthorizationReader{}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)

		_, execErr := svc.Execute(context.Background(), domain.PermissionCheckRequest{
			Namespace: "app",
			Object:    "resource-1",
			Relation:  "view",
		})

		assert.ErrorIs(t, execErr, commondomain.ErrUnauthorized)
		assert.False(t, keto.called)
	})

	t.Run("validates required fields", func(t *testing.T) {
		keto := &mockAuthorizationReader{}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)

		_, execErr := svc.Execute(authCtx, domain.PermissionCheckRequest{
			Namespace: "",
			Object:    "resource-1",
			Relation:  "view",
		})

		assert.ErrorIs(t, execErr, commondomain.ErrRequestInvalid)
		assert.False(t, keto.called)
	})

	t.Run("enforces public allowlist", func(t *testing.T) {
		keto := &mockAuthorizationReader{}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)

		_, execErr := svc.Execute(authCtx, domain.PermissionCheckRequest{
			Namespace: "app",
			Object:    "resource-1",
			Relation:  "edit",
		})

		assert.ErrorIs(t, execErr, commondomain.ErrForbidden)
		assert.False(t, keto.called)
	})

	t.Run("maps keto errors to authz unavailable", func(t *testing.T) {
		keto := &mockAuthorizationReader{err: errors.New("keto unavailable")}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)

		_, execErr := svc.Execute(authCtx, domain.PermissionCheckRequest{
			Namespace: "app",
			Object:    "resource-1",
			Relation:  "view",
		})

		assert.ErrorIs(t, execErr, commondomain.ErrAuthzUnavailable)
		assert.Contains(t, execErr.Error(), "keto unavailable")
		assert.True(t, keto.called)
	})

	t.Run("returns permission decision from keto", func(t *testing.T) {
		keto := &mockAuthorizationReader{allowed: true}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)

		allowed, execErr := svc.Execute(authCtx, domain.PermissionCheckRequest{
			Namespace: "app",
			Object:    "resource-1",
			Relation:  "view",
		})

		require.NoError(t, execErr)
		assert.True(t, allowed)
		assert.True(t, keto.called)
		assert.Equal(t, "user-1", keto.subject.ID)
	})

	t.Run("returns authz unavailable when claims prefetch failed", func(t *testing.T) {
		keto := &mockAuthorizationReader{}
		svc := domain.NewPublicPermissionCheck(keto, allowlist)
		ctx := commonroles.WithClaims(context.Background(), commonroles.Claims{
			Subject:       "user-1",
			Authenticated: true,
			Err:           errors.New("keto down"),
		})

		_, execErr := svc.Execute(ctx, domain.PermissionCheckRequest{
			Namespace: "app",
			Object:    "resource-1",
			Relation:  "view",
		})

		assert.ErrorIs(t, execErr, commondomain.ErrAuthzUnavailable)
		assert.False(t, keto.called)
	})
}

package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type proxyAdminAuthorizationReader struct {
	allowed   bool
	err       error
	called    bool
	namespace string
	object    string
	relation  string
	subject   ketoclient.Subject
}

func (m *proxyAdminAuthorizationReader) CheckPermission(
	ctx context.Context,
	namespace string,
	object string,
	relation string,
	subject ketoclient.Subject,
) (bool, error) {
	m.called = true
	m.namespace = namespace
	m.object = object
	m.relation = relation
	m.subject = subject
	return m.allowed, m.err
}

func TestProxyAdminCheck_Execute(t *testing.T) {
	t.Run("checks only fixed Tadoku admin membership", func(t *testing.T) {
		keto := &proxyAdminAuthorizationReader{allowed: true}
		svc := domain.NewProxyAdminCheck(keto)

		allowed, err := svc.Execute(context.Background(), "6f75df3f-a162-4c1a-b206-5ec2c62b25a9")

		require.NoError(t, err)
		assert.True(t, allowed)
		assert.True(t, keto.called)
		assert.Equal(t, "app", keto.namespace)
		assert.Equal(t, "tadoku", keto.object)
		assert.Equal(t, "admins", keto.relation)
		assert.Equal(t, ketoclient.Subject{ID: "6f75df3f-a162-4c1a-b206-5ec2c62b25a9"}, keto.subject)
	})

	t.Run("returns denied decision", func(t *testing.T) {
		keto := &proxyAdminAuthorizationReader{allowed: false}
		svc := domain.NewProxyAdminCheck(keto)

		allowed, err := svc.Execute(context.Background(), "d1fc57e1-1654-4092-a044-15db9f894dd9")

		require.NoError(t, err)
		assert.False(t, allowed)
		assert.True(t, keto.called)
	})

	t.Run("rejects an empty subject without calling Keto", func(t *testing.T) {
		keto := &proxyAdminAuthorizationReader{}
		svc := domain.NewProxyAdminCheck(keto)

		_, err := svc.Execute(context.Background(), "")

		assert.ErrorIs(t, err, commondomain.ErrRequestInvalid)
		assert.False(t, keto.called)
	})

	t.Run("rejects a malformed Kratos subject without calling Keto", func(t *testing.T) {
		keto := &proxyAdminAuthorizationReader{}
		svc := domain.NewProxyAdminCheck(keto)

		_, err := svc.Execute(context.Background(), "forged-subject")

		assert.ErrorIs(t, err, commondomain.ErrRequestInvalid)
		assert.False(t, keto.called)
	})

	t.Run("maps Keto errors to authorization unavailable", func(t *testing.T) {
		ketoErr := errors.New("keto unavailable")
		keto := &proxyAdminAuthorizationReader{err: ketoErr}
		svc := domain.NewProxyAdminCheck(keto)

		_, err := svc.Execute(context.Background(), "723f68ec-2260-4c93-8a54-0443629a7285")

		assert.ErrorIs(t, err, commondomain.ErrAuthzUnavailable)
		assert.ErrorIs(t, err, ketoErr)
		assert.True(t, keto.called)
	})
}

package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

const (
	proxyAdminNamespace = "app"
	proxyAdminObject    = "tadoku"
	proxyAdminRelation  = "admins"
)

type ProxyAdminPermissionReader interface {
	CheckPermission(
		ctx context.Context,
		namespace string,
		object string,
		relation string,
		subject ketoclient.Subject,
	) (bool, error)
}

type ProxyAdminCheck struct {
	keto ProxyAdminPermissionReader
}

func NewProxyAdminCheck(keto ProxyAdminPermissionReader) *ProxyAdminCheck {
	return &ProxyAdminCheck{keto: keto}
}

func (s *ProxyAdminCheck) Execute(ctx context.Context, subjectID string) (bool, error) {
	subject, err := uuid.Parse(subjectID)
	if err != nil || subject == uuid.Nil {
		return false, fmt.Errorf("%w: subject must be a UUID", commondomain.ErrRequestInvalid)
	}

	allowed, err := s.keto.CheckPermission(
		ctx,
		proxyAdminNamespace,
		proxyAdminObject,
		proxyAdminRelation,
		ketoclient.Subject{ID: subjectID},
	)
	if err != nil {
		return false, fmt.Errorf("%w: check proxy admin permission failed: %w", commondomain.ErrAuthzUnavailable, err)
	}
	return allowed, nil
}

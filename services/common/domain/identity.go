package domain

import (
	"context"
	"time"
)

type Identity interface {
	// GetSubject returns the token "sub" (subject) claim, which identifies
	// the principal the token was issued for. It is stable and unique within
	// the issuer, and is the primary identifier for authorization.
	GetSubject() string
}

type UserIdentity struct {
	// Subject is the token "sub" (subject) claim. For user tokens this is the
	// stable unique user ID from the identity provider (Kratos), and is the
	// primary identifier we use for authorization and role lookups.
	Subject     string
	DisplayName string
	Email       string
	CreatedAt   time.Time
}

func (u *UserIdentity) GetSubject() string { return u.Subject }

type ServiceIdentity struct {
	// Subject is the token "sub" (subject) claim. For service tokens this is
	// the full Kubernetes service account name in the form:
	// "system:serviceaccount:<namespace>:<name>".
	Subject   string
	Name      string
	Namespace string
	Audience  []string
}

func (s *ServiceIdentity) GetSubject() string { return s.Subject }

func ParseIdentity(ctx context.Context) Identity {
	if identity, ok := ctx.Value(CtxIdentityKey).(Identity); ok && identity != nil {
		return identity
	}

	return nil
}

func ParseUserIdentity(ctx context.Context) *UserIdentity {
	if identity := ParseIdentity(ctx); identity != nil {
		if user, ok := identity.(*UserIdentity); ok {
			return user
		}
	}

	return nil
}

func ParseServiceIdentity(ctx context.Context) *ServiceIdentity {
	if identity := ParseIdentity(ctx); identity != nil {
		if svc, ok := identity.(*ServiceIdentity); ok {
			return svc
		}
	}

	return nil
}

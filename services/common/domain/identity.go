package domain

import (
	"context"
	"time"
)

// IdentityType distinguishes between user and service identities.
type IdentityType string

const (
	IdentityTypeUser    IdentityType = "user"
	IdentityTypeService IdentityType = "service"
)

// Identity is the common interface for all identity types.
type Identity interface {
	// GetSubject returns the unique subject identifier from the token.
	GetSubject() string
	GetType() IdentityType
	IsUser() bool
	IsService() bool
}

// UserIdentity represents a human user authenticated via Kratos.
type UserIdentity struct {
	// Subject is the token "sub" claim (unique user ID).
	Subject     string
	DisplayName string
	Email       string
	Role        Role
	CreatedAt   time.Time
}

func (u *UserIdentity) GetSubject() string    { return u.Subject }
func (u *UserIdentity) GetType() IdentityType { return IdentityTypeUser }
func (u *UserIdentity) IsUser() bool          { return true }
func (u *UserIdentity) IsService() bool       { return false }

// ServiceIdentity represents a service authenticated via K8s SA.
type ServiceIdentity struct {
	// Subject is the token "sub" claim (full service account name).
	Subject   string
	Name      string
	Namespace string
	Audience  []string
}

func (s *ServiceIdentity) GetSubject() string    { return s.Subject }
func (s *ServiceIdentity) GetType() IdentityType { return IdentityTypeService }
func (s *ServiceIdentity) IsUser() bool          { return false }
func (s *ServiceIdentity) IsService() bool       { return true }

// ParseIdentity extracts the identity from context.
func ParseIdentity(ctx context.Context) Identity {
	if identity, ok := ctx.Value(CtxIdentityKey).(Identity); ok && identity != nil {
		return identity
	}

	if session, ok := ctx.Value(CtxSessionKey).(*UserIdentity); ok && session != nil {
		return session
	}

	return nil
}

// ParseUserIdentity extracts user identity, returns nil if not a user.
func ParseUserIdentity(ctx context.Context) *UserIdentity {
	if identity := ParseIdentity(ctx); identity != nil {
		if user, ok := identity.(*UserIdentity); ok {
			return user
		}
	}

	return nil
}

// ParseServiceIdentity extracts service identity, returns nil if not a service.
func ParseServiceIdentity(ctx context.Context) *ServiceIdentity {
	if identity := ParseIdentity(ctx); identity != nil {
		if svc, ok := identity.(*ServiceIdentity); ok {
			return svc
		}
	}

	return nil
}

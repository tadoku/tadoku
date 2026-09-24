package domain

import (
	"context"
	"time"
)

type IdentityType string

const (
	IdentityTypeUser    IdentityType = "user"
	IdentityTypeService IdentityType = "service"
)

type Identity interface {
	GetSubject() string
	GetType() IdentityType
	IsUser() bool
	IsService() bool
}

type UserIdentity struct {
	Subject     string
	DisplayName string
	Email       string
	CreatedAt   time.Time
}

func (u *UserIdentity) GetSubject() string    { return u.Subject }
func (u *UserIdentity) GetType() IdentityType { return IdentityTypeUser }
func (u *UserIdentity) IsUser() bool          { return true }
func (u *UserIdentity) IsService() bool       { return false }

type ServiceIdentity struct {
	Subject   string
	Name      string
	Namespace string
	Audience  []string
}

func (s *ServiceIdentity) GetSubject() string    { return s.Subject }
func (s *ServiceIdentity) GetType() IdentityType { return IdentityTypeService }
func (s *ServiceIdentity) IsUser() bool          { return false }
func (s *ServiceIdentity) IsService() bool       { return true }

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

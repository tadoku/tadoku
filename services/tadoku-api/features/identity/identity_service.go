package identity

import (
	"context"

	"github.com/google/uuid"
)

type Service struct{}

var _ Accounts = (*Service)(nil)

func NewService() *Service { return &Service{} }
func (*Service) FindAccount(context.Context, uuid.UUID) (Account, error) {
	return Account{}, ErrNotImplemented
}
func (*Service) UserExists(context.Context, uuid.UUID) (bool, error) { return false, ErrNotImplemented }
func (*Service) ListAccounts(context.Context, ListAccounts) (AccountPage, error) {
	return AccountPage{}, ErrNotImplemented
}
func (*Service) Register(context.Context, Registration) (LoginResult, error) {
	return LoginResult{}, ErrNotImplemented
}
func (*Service) Login(context.Context, Login) (LoginResult, error) {
	return LoginResult{}, ErrNotImplemented
}
func (*Service) Logout(context.Context, string) (Cookie, error) { return Cookie{}, ErrNotImplemented }
func (*Service) Session(context.Context, string) (SessionResult, error) {
	return SessionResult{}, ErrNotImplemented
}
func (*Service) CheckSession(context.Context, string) (Session, error) { return nil, ErrNotImplemented }
func (*Service) RequestRecovery(context.Context, Recovery) (RecoveryResult, error) {
	return RecoveryResult{}, ErrNotImplemented
}
func (*Service) ConsumeRecovery(context.Context, string) (RecoveryLogin, error) {
	return RecoveryLogin{}, ErrNotImplemented
}
func (*Service) UpdateProfile(context.Context, string, ProfileSettings) (ProfileResult, error) {
	return ProfileResult{}, ErrNotImplemented
}
func (*Service) UpdatePassword(context.Context, string, string) error { return ErrNotImplemented }

func (*Service) DeleteAccount(context.Context, uuid.UUID) error { return ErrNotImplemented }

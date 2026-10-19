package identity

import (
	"context"
	"errors"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfile requires the application's business transaction, including notices and enqueue.
func (s *Service) UpdateProfile(
	ctx context.Context,
	header CookieHeader,
	settings ProfileSettings,
) (ProfileResult, error) {
	session, err := s.lockSettingsSession(ctx, header)
	if err != nil {
		return ProfileResult{}, err
	}
	if settings.Email == nil && settings.DisplayName == nil {
		return ProfileResult{}, errx.NewInvalidInputError("email or display_name is required")
	}

	account := session.Account
	var messages []mail.Message
	if settings.Email != nil {
		email, err := ParseEmail(*settings.Email)
		if err != nil {
			return ProfileResult{}, err
		}
		if email != account.Email {
			if err := recentSession(session); err != nil {
				return ProfileResult{}, err
			}
			if err := s.repository.updateEmail(ctx, account.ID, email, timex.Now()); err != nil {
				return ProfileResult{}, err
			}
			if err := s.repository.invalidateRecoveryTokens(ctx, account.ID, timex.Now()); err != nil {
				return ProfileResult{}, err
			}
			messages = []mail.Message{{
				Template:  mail.AccountEmailChanged,
				Recipient: account.Email.value,
				Params: map[string]string{
					"new_email":  email.value,
					"changed_at": timex.Now().Format(time.RFC3339),
				},
			}}
			account.Email = email
		}
	}
	if settings.DisplayName != nil {
		name, err := ParseDisplayName(*settings.DisplayName)
		if err != nil {
			return ProfileResult{}, err
		}
		if err := s.repository.updateDisplayName(ctx, account.ID, name, timex.Now()); err != nil {
			return ProfileResult{}, err
		}
		account.DisplayName = name
	}
	return ProfileResult{Account: account, Emails: messages}, nil
}

// UpdatePassword requires the application's transaction through credential change and revocation.
func (s *Service) UpdatePassword(ctx context.Context, header CookieHeader, raw string) error {
	session, err := s.lockSettingsSession(ctx, header)
	if err != nil {
		return err
	}
	if err := recentSession(session); err != nil {
		return err
	}
	password, err := ParsePassword(raw, session.Account.Email)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password.value), 12)
	if err != nil {
		return err
	}
	if err := s.repository.updatePassword(ctx, session.Account.ID, string(hash), timex.Now()); err != nil {
		return err
	}
	if err := s.repository.invalidateRecoveryTokens(ctx, session.Account.ID, timex.Now()); err != nil {
		return err
	}
	return s.repository.revokeOtherSessions(ctx, session.Account.ID, hashToken(cookieValue(header, s.cookie(ctx).Name)))
}

func (s *Service) lockSettingsSession(ctx context.Context, header CookieHeader) (AuthenticatedSession, error) {
	result, err := s.Session(ctx, header)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	if result.AccountApp == V1 {
		return AuthenticatedSession{}, ErrNotImplemented
	}
	session, ok := result.Session.(AuthenticatedSession)
	if !ok {
		return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
	}
	if _, err := s.repository.lockAccount(ctx, session.Account.ID); err != nil {
		if errors.Is(err, errAccountNotFound) {
			return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
		}
		return AuthenticatedSession{}, err
	}

	current, err := s.repository.session(ctx, hashToken(cookieValue(header, s.cookie(ctx).Name)), timex.Now())
	if err != nil {
		return AuthenticatedSession{}, err
	}
	session, ok = current.(AuthenticatedSession)
	if !ok {
		return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
	}
	return session, nil
}

func recentSession(session AuthenticatedSession) error {
	age := timex.Now().Sub(session.AuthenticatedAt)
	if age < 0 || age > time.Hour {
		return ErrStaleSession
	}
	return nil
}

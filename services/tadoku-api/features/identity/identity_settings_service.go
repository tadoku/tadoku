package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
		if email.value != account.Email.recipient() {
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
				Recipient: account.Email.recipient(),
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

type PasswordUpdate struct {
	header   CookieHeader
	password Password
	hash     string
}

// HashPasswordUpdate must run before opening the application's business transaction.
func (s *Service) HashPasswordUpdate(ctx context.Context, header CookieHeader, raw string) (PasswordUpdate, error) {
	session, err := s.SettingsSession(ctx, header)
	if err != nil {
		return PasswordUpdate{}, err
	}
	if err := recentSession(session); err != nil {
		return PasswordUpdate{}, err
	}
	password, err := ParsePassword(raw, session.Account.Email)
	if err != nil {
		return PasswordUpdate{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password.value), 12)
	if err != nil {
		return PasswordUpdate{}, err
	}
	return PasswordUpdate{header: header, password: password, hash: string(hash)}, nil
}

// ApplyPasswordUpdate requires the application's transaction through credential change and revocation.
func (s *Service) ApplyPasswordUpdate(ctx context.Context, update PasswordUpdate) error {
	session, err := s.lockSettingsSession(ctx, update.header)
	if err != nil {
		return err
	}
	if err := recentSession(session); err != nil {
		return err
	}
	if _, err := ParsePassword(update.password.value, session.Account.Email); err != nil {
		return err
	}
	if update.hash == "" {
		return errx.NewInvalidInputError("password update must be hashed before applying")
	}

	if err := s.repository.updatePassword(ctx, session.Account.ID, update.hash, timex.Now()); err != nil {
		return err
	}
	if err := s.repository.invalidateRecoveryTokens(ctx, session.Account.ID, timex.Now()); err != nil {
		return err
	}
	return s.repository.revokeOtherSessions(
		ctx,
		session.Account.ID,
		hashToken(cookieValue(update.header, s.cookie(ctx).Name)),
	)
}

type AccountDeletion struct {
	header CookieHeader
	email  Email
}

// AdmitAccountDeletion must run before the application's business transaction.
func (s *Service) AdmitAccountDeletion(ctx context.Context, header CookieHeader, email string) (AccountDeletion, error) {
	session, err := s.SettingsSession(ctx, header)
	if err != nil {
		return AccountDeletion{}, err
	}
	if err := recentSession(session); err != nil {
		return AccountDeletion{}, err
	}
	address, err := ParseEmail(email)
	if err != nil {
		return AccountDeletion{}, err
	}
	return AccountDeletion{header: header, email: address}, nil
}

// RequestAccountDeletion requires the transaction that files the deletion for the returned user. It locks the
// account row before filing locks the users row.
func (s *Service) RequestAccountDeletion(ctx context.Context, deletion AccountDeletion) (uuid.UUID, Cookie, error) {
	session, err := s.lockSettingsSession(ctx, deletion.header)
	if err != nil {
		return uuid.Nil, Cookie{}, err
	}
	if err := recentSession(session); err != nil {
		return uuid.Nil, Cookie{}, err
	}

	matches, err := s.repository.accountHasEmail(ctx, session.Account.ID, deletion.email.value)
	if err != nil {
		return uuid.Nil, Cookie{}, err
	}
	if !matches {
		return uuid.Nil, Cookie{}, ErrConfirmationMismatch
	}
	return session.Account.ID, s.clearedCookie(ctx), nil
}

func (s *Service) SettingsSession(ctx context.Context, header CookieHeader) (AuthenticatedSession, error) {
	current, err := s.Session(ctx, header)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	session, ok := current.(AuthenticatedSession)
	if !ok {
		return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
	}
	return session, nil
}

func (s *Service) lockSettingsSession(ctx context.Context, header CookieHeader) (AuthenticatedSession, error) {
	session, err := s.SettingsSession(ctx, header)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	if _, err := usableAccount(s.repository.lockAccount(ctx, session.Account.ID)); err != nil {
		if errors.Is(err, errAccountNotFound) {
			return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
		}
		return AuthenticatedSession{}, err
	}

	current, err := s.session(ctx, cookieValue(header, s.cookie(ctx).Name))
	if err != nil {
		return AuthenticatedSession{}, err
	}
	session, ok := current.(AuthenticatedSession)
	if !ok {
		return AuthenticatedSession{}, errx.NewUnauthorizedError("a valid session is required")
	}
	return session, nil
}

func recentSession(session AuthenticatedSession) error {
	if !session.RecentLogin() {
		return ErrStaleSession
	}
	return nil
}

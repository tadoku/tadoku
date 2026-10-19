package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository   *Repository
	throttle     *Throttle
	legacy       *LegacySessions
	flags        *featureflags.Evaluator
	cookieDomain string
}

var _ Accounts = (*Service)(nil)

func NewService(
	repository *Repository,
	throttle *Throttle,
	legacy *LegacySessions,
	flags *featureflags.Evaluator,
	cookieDomain string,
) *Service {
	return &Service{
		repository:   repository,
		throttle:     throttle,
		legacy:       legacy,
		flags:        flags,
		cookieDomain: cookieDomain,
	}
}

func (*Service) FindAccount(context.Context, uuid.UUID) (Account, error) {
	return Account{}, ErrNotImplemented
}

func (*Service) UserExists(context.Context, uuid.UUID) (bool, error) { return false, ErrNotImplemented }

func (*Service) ListAccounts(context.Context, ListAccounts) (AccountPage, error) {
	return AccountPage{}, ErrNotImplemented
}

func (s *Service) Register(ctx context.Context, registration Registration) (LoginResult, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	if app == V1 {
		return LoginResult{}, ErrNotImplemented
	}
	email, err := ParseEmail(registration.Email)
	if err != nil {
		return LoginResult{}, err
	}
	name, err := ParseDisplayName(registration.DisplayName)
	if err != nil {
		return LoginResult{}, err
	}
	password, err := ParsePassword(registration.Password, email)
	if err != nil {
		return LoginResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password.value), 12)
	if err != nil {
		return LoginResult{}, err
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return LoginResult{}, err
	}
	account := Account{ID: id, Email: email, DisplayName: name, CreatedAt: timex.Now()}
	result, err := s.newLogin(ctx, account)
	if err != nil {
		return LoginResult{}, err
	}

	if err := s.repository.register(ctx, result, string(hash)); err != nil {
		return LoginResult{}, err
	}
	return result, nil
}

func (s *Service) Login(ctx context.Context, login Login) (LoginResult, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return LoginResult{}, err
	}
	if app == V1 {
		return LoginResult{}, ErrNotImplemented
	}
	email, err := ParseEmail(login.Email)
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.throttle.allow(ctx, email, login.IP); err != nil {
		return LoginResult{}, err
	}

	credential, err := s.repository.credentials(ctx, email)
	if err != nil && !errors.Is(err, errAccountNotFound) {
		return LoginResult{}, err
	}
	found := err == nil
	hash := credential.passwordHash
	if !found {
		hash = dummyPasswordHash
	}
	password := login.Password
	if len(password) > 72 {
		password = "oversized password comparison"
	}

	comparison := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if comparison != nil || !found || !credential.active || len(login.Password) > 72 {
		return LoginResult{}, ErrWrongPassword
	}

	return s.loginSession(ctx, credential.account)
}

func (s *Service) Logout(ctx context.Context, header CookieHeader) (Cookie, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return Cookie{}, err
	}

	cookie := s.cookie(ctx)
	if app == V1 {
		cookie.Name = "ory_kratos_session"
		value := cookieValue(header, cookie.Name)
		if value != "" {
			session, err := s.legacy.Check(ctx, cookie.Name+"="+value)
			if err != nil {
				return Cookie{}, errx.NewUnavailableError("identity provider unavailable", err)
			}
			if session != nil {
				if err := s.legacy.Revoke(ctx, session.Id); err != nil {
					return Cookie{}, errx.NewUnavailableError("identity provider unavailable", err)
				}
			}
		}
	} else if value := cookieValue(header, cookie.Name); validSessionToken(value) {
		if err := s.repository.revokeSession(ctx, hashToken(value)); err != nil {
			return Cookie{}, err
		}
	}

	cookie.MaxAge, cookie.Expires = -1, time.Unix(1, 0).UTC()
	return cookie, nil
}

func (s *Service) Session(ctx context.Context, header CookieHeader) (SessionResult, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return SessionResult{}, err
	}
	var session Session = GuestSession{}
	if app == V1 {
		session, err = s.legacySession(ctx, header)
	} else if value := cookieValue(header, s.cookie(ctx).Name); validSessionToken(value) {
		session, err = s.repository.session(ctx, hashToken(value), timex.Now())
	}
	return SessionResult{AccountApp: app, Session: session}, err
}

func (s *Service) CheckSession(ctx context.Context, header CookieHeader) (Session, error) {
	result, err := s.Session(ctx, header)
	return result.Session, err
}

func (*Service) RequestRecovery(context.Context, Recovery) (RecoveryResult, error) {
	return RecoveryResult{}, ErrNotImplemented
}

func (*Service) ConsumeRecovery(context.Context, RecoveryToken) (RecoveryLogin, error) {
	return RecoveryLogin{}, ErrNotImplemented
}

func (*Service) UpdateProfile(context.Context, CookieHeader, ProfileSettings) (ProfileResult, error) {
	return ProfileResult{}, ErrNotImplemented
}

func (*Service) UpdatePassword(context.Context, CookieHeader, string) error { return ErrNotImplemented }

func (*Service) DeleteAccount(context.Context, uuid.UUID) error { return ErrNotImplemented }

const dummyPasswordHash = "$2b$12$hady7Vf2de/ngyT9TV6N/usj/9yeHugOWtf8eCYGfnWNO.z9CGeeu"

func (s *Service) accountApp(ctx context.Context) (AccountApp, error) {
	if _, ok := tenant.FromContext(ctx); !ok {
		return "", errx.NewUnavailableError("identity tenant is required", nil)
	}
	if s.flags.TenantBoolean(ctx, featureflags.IdentityV2) {
		return V2, nil
	}
	return V1, nil
}

func (s *Service) cookie(ctx context.Context) Cookie {
	key, _ := tenant.FromContext(ctx)
	name := "tadoku_session"
	lifetime := 720 * time.Hour
	if key != tenant.Production() {
		name += "_" + strings.ReplaceAll(key.String(), "/", "_")
		lifetime = 7 * 24 * time.Hour
	}
	return Cookie{
		Name:    name,
		Domain:  s.cookieDomain,
		Expires: timex.Now().Add(lifetime),
		MaxAge:  int(lifetime.Seconds()),
	}
}

func (s *Service) loginSession(ctx context.Context, account Account) (LoginResult, error) {
	result, err := s.newLogin(ctx, account)
	if err != nil {
		return LoginResult{}, err
	}
	if err := s.repository.createSession(ctx, result.Session, hashToken(result.Cookie.Value)); err != nil {
		return LoginResult{}, err
	}
	return result, nil
}

func (s *Service) newLogin(ctx context.Context, account Account) (LoginResult, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return LoginResult{}, err
	}
	cookie := s.cookie(ctx)
	cookie.Value = base64.RawURLEncoding.EncodeToString(secret)
	session := AuthenticatedSession{
		Account:         account,
		AuthenticatedAt: timex.Now(),
		ExpiresAt:       cookie.Expires,
	}
	return LoginResult{Session: session, Cookie: cookie}, nil
}

func cookieValue(header CookieHeader, name string) string {
	request := &http.Request{Header: http.Header{"Cookie": []string{string(header)}}}
	value, count := "", 0
	for _, cookie := range request.Cookies() {
		if cookie.Name == name {
			value, count = cookie.Value, count+1
		}
	}
	if count != 1 {
		return ""
	}
	return value
}

func validSessionToken(raw string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	return err == nil && len(decoded) == 32 && base64.RawURLEncoding.EncodeToString(decoded) == raw
}

func (s *Service) legacySession(ctx context.Context, header CookieHeader) (Session, error) {
	value := cookieValue(header, "ory_kratos_session")
	if value == "" {
		return GuestSession{}, nil
	}
	session, err := s.legacy.Check(ctx, "ory_kratos_session="+value)
	if err != nil {
		return nil, errx.NewUnavailableError("identity provider unavailable", err)
	}
	if session == nil {
		return GuestSession{}, nil
	}
	id, err := uuid.Parse(session.Identity.Id)
	if err != nil {
		return nil, errx.NewUnavailableError("identity provider returned an invalid account", nil)
	}
	traits, ok := session.Identity.Traits.(map[string]interface{})
	if !ok {
		return nil, errx.NewUnavailableError("identity provider returned invalid traits", nil)
	}
	email, emailOK := traits["email"].(string)
	name, nameOK := traits["display_name"].(string)
	if !emailOK || !nameOK {
		return nil, errx.NewUnavailableError("identity provider returned invalid traits", nil)
	}
	return AuthenticatedSession{
		Account: Account{
			ID:          id,
			Email:       Email{email},
			DisplayName: DisplayName{name},
			CreatedAt:   session.Identity.GetCreatedAt(),
		},
		AuthenticatedAt: session.GetAuthenticatedAt(),
		ExpiresAt:       session.GetExpiresAt(),
	}, nil
}

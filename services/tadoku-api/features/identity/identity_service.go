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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository   *Repository
	throttle     *Throttle
	cookieDomain string
}

func NewService(repository *Repository, throttle *Throttle, cookieDomain string) *Service {
	return &Service{repository: repository, throttle: throttle, cookieDomain: cookieDomain}
}

func (s *Service) Register(ctx context.Context, registration Registration) (LoginResult, error) {
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
	if err := s.throttle.admitRegistration(ctx, registration.IP); err != nil {
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

	return s.loginSession(ctx, credential)
}

func (s *Service) Logout(ctx context.Context, header CookieHeader) (Cookie, error) {
	cookie := s.cookie(ctx)
	if value := cookieValue(header, cookie.Name); validSessionToken(value) {
		if err := s.repository.revokeSession(ctx, hashToken(value)); err != nil {
			return Cookie{}, err
		}
	}

	cookie.MaxAge, cookie.Expires = -1, time.Unix(1, 0).UTC()
	return cookie, nil
}

func (s *Service) Session(ctx context.Context, header CookieHeader) (Session, error) {
	value := cookieValue(header, s.cookie(ctx).Name)
	if !validSessionToken(value) {
		return GuestSession{}, nil
	}
	return s.repository.session(ctx, hashToken(value), timex.Now())
}

const dummyPasswordHash = "$2b$12$hady7Vf2de/ngyT9TV6N/usj/9yeHugOWtf8eCYGfnWNO.z9CGeeu"

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

func (s *Service) loginSession(ctx context.Context, credential credentials) (LoginResult, error) {
	result, err := s.newLogin(ctx, credential.account)
	if err != nil {
		return LoginResult{}, err
	}
	err = s.repository.createSession(
		ctx,
		result.Session,
		hashToken(result.Cookie.Value),
		credential.passwordHash,
	)
	if err != nil {
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

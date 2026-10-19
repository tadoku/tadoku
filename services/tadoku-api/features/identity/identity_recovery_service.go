package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

// RequestRecovery requires the application's business transaction, including email writes and enqueue.
func (s *Service) RequestRecovery(ctx context.Context, recovery Recovery) (RecoveryResult, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return RecoveryResult{}, err
	}
	if app == V1 {
		return RecoveryResult{}, ErrNotImplemented
	}
	email, err := ParseEmail(recovery.Email)
	if err != nil {
		return RecoveryResult{}, err
	}
	if err := ValidateReturnTo(recovery.ReturnTo); err != nil {
		return RecoveryResult{}, err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return RecoveryResult{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	origin := strings.TrimSuffix(s.appURL(ctx, ""), "/")
	if err := ValidateOrigin(origin); err != nil {
		return RecoveryResult{}, err
	}
	link := origin + "/api/internal/identity/recovery/" + token +
		"?return_to=" + url.QueryEscape(recovery.ReturnTo)
	if len(link) > 4096 {
		return RecoveryResult{}, &Error{
			"invalid_return_to", errx.NewInvalidInputError("return_to makes the recovery link exceed 4096 bytes"),
		}
	}

	credential, err := s.repository.lockRecoveryAccount(ctx, email)
	if errors.Is(err, errAccountNotFound) {
		return RecoveryResult{}, nil
	}
	if err != nil {
		return RecoveryResult{}, err
	}
	err = s.repository.createRecoveryToken(ctx, credential.account.ID, hashToken(token), timex.Now().Add(time.Hour))
	if err != nil {
		return RecoveryResult{}, err
	}
	return RecoveryResult{Emails: []mail.Message{{
		Template:  mail.AccountRecovery,
		Recipient: credential.account.Email.value,
		Params:    map[string]string{"recovery_url": link},
	}}}, nil
}

// ConsumeRecovery requires the application's business transaction through the new session write.
func (s *Service) ConsumeRecovery(ctx context.Context, raw RecoveryToken) (RecoveryLogin, error) {
	app, err := s.accountApp(ctx)
	if err != nil {
		return RecoveryLogin{}, err
	}
	if app == V1 {
		return RecoveryLogin{}, ErrNotImplemented
	}
	if !validSessionToken(string(raw)) {
		return RecoveryLogin{}, ErrInvalidRecoveryToken
	}
	hash := hashToken(string(raw))
	id, err := s.repository.recoveryAccount(ctx, hash)
	if err != nil {
		return RecoveryLogin{}, err
	}
	credential, err := s.repository.lockAccount(ctx, id)
	if errors.Is(err, errAccountNotFound) {
		return RecoveryLogin{}, ErrInvalidRecoveryToken
	}
	if err != nil {
		return RecoveryLogin{}, err
	}
	if err := s.repository.useRecoveryToken(ctx, hash, timex.Now()); err != nil {
		return RecoveryLogin{}, err
	}
	if err := s.repository.invalidateRecoveryTokens(ctx, id, timex.Now()); err != nil {
		return RecoveryLogin{}, err
	}
	if err := s.repository.revokeOtherSessions(ctx, id, ""); err != nil {
		return RecoveryLogin{}, err
	}

	login, err := s.loginSession(ctx, credential)
	if err != nil {
		return RecoveryLogin{}, err
	}
	return RecoveryLogin{Login: login, RedirectTo: s.appURL(ctx, "authv2")}, nil
}

func (s *Service) appURL(ctx context.Context, app string) string {
	key, _ := tenant.FromContext(ctx)
	host := s.cookieDomain
	if app != "" {
		host = app + "." + host
	}
	if key != tenant.Production() {
		route := strings.Split(key.String(), "/")[1]
		if s.cookieDomain == "tadoku.app" {
			host = "preview.tadoku.app"
			if app != "" {
				host = app + "." + host
			}
		}
		host = route + "." + host
	}
	return "https://" + host + "/"
}

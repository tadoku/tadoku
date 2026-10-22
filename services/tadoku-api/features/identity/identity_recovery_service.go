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

type RecoveryRequest struct {
	email Email
	link  string
	token string
}

// AdmitRecovery must run before the application's business transaction. A nil request without an error
// means admission suppressed the recipient; answer exactly as for an unknown address.
func (s *Service) AdmitRecovery(ctx context.Context, recovery Recovery) (*RecoveryRequest, error) {
	email, err := ParseEmail(recovery.Email)
	if err != nil {
		return nil, err
	}
	if _, err := s.RecoverySettingsURL(ctx, &recovery.ReturnTo); err != nil {
		return nil, err
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(secret)
	origin := strings.TrimSuffix(s.appURL(ctx, ""), "/")
	if err := ValidateOrigin(origin); err != nil {
		return nil, err
	}
	link := origin + "/api/internal/identity/recovery/" + token +
		"?return_to=" + url.QueryEscape(recovery.ReturnTo)
	if len(link) > 4096 {
		return nil, &Error{
			"invalid_return_to", errx.NewInvalidInputError("return_to makes the recovery link exceed 4096 bytes"),
		}
	}

	admitted, err := s.throttle.admitRecovery(ctx, email, recovery.IP)
	if err != nil || !admitted {
		return nil, err
	}
	return &RecoveryRequest{email: email, link: link, token: token}, nil
}

// WriteRecovery requires the application's business transaction, including email writes and enqueue.
func (s *Service) WriteRecovery(ctx context.Context, request RecoveryRequest) (RecoveryResult, error) {
	credential, err := s.repository.lockRecoveryAccount(ctx, request.email)
	if errors.Is(err, errAccountNotFound) {
		return RecoveryResult{}, nil
	}
	if err != nil {
		return RecoveryResult{}, err
	}
	err = s.repository.createRecoveryToken(
		ctx,
		credential.account.ID,
		hashToken(request.token),
		timex.Now().Add(time.Hour),
	)
	if err != nil {
		return RecoveryResult{}, err
	}
	return RecoveryResult{Emails: []mail.Message{{
		Template:  mail.AccountRecovery,
		Recipient: credential.account.Email.recipient(),
		Params:    map[string]string{"recovery_url": request.link},
	}}}, nil
}

// RequestRecovery requires the application's business transaction, including email writes and enqueue.
// HTTP callers admit with AdmitRecovery before the transaction and use WriteRecovery inside it.
func (s *Service) RequestRecovery(ctx context.Context, recovery Recovery) (RecoveryResult, error) {
	request, err := s.AdmitRecovery(ctx, recovery)
	if err != nil || request == nil {
		return RecoveryResult{}, err
	}
	return s.WriteRecovery(ctx, *request)
}

// ConsumeRecovery requires the application's business transaction through the new session write.
func (s *Service) ConsumeRecovery(ctx context.Context, raw RecoveryToken) (RecoveryLogin, error) {
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

func (s *Service) RecoverySettingsURL(ctx context.Context, returnTo *string) (string, error) {
	location := s.appURL(ctx, "authv2")
	if returnTo != nil {
		if err := ValidateReturnTo(*returnTo); err != nil {
			return "", err
		}
		location += "?return_to=" + url.QueryEscape(*returnTo)
	}
	if err := ValidateReturnTo(location); err != nil {
		return "", &Error{
			"invalid_return_to",
			errx.NewInvalidInputError("return_to makes the Settings URL invalid or exceed 2048 bytes"),
		}
	}
	return location, nil
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

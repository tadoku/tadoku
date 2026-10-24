package http

import (
	"context"
	"errors"
	"mime"
	"net"
	stdhttp "net/http"
	"net/mail"
	"net/netip"
	"strings"

	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	callbackopenapi "github.com/tadoku/tadoku/services/tadoku-api/generated/openapi/callback"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type identityRequestKey struct{}
type identityRequestMetadata struct {
	cookies app.IdentityCookieHeader
	ip      string
}

func (s *server) withIdentityRequest(next openapi.StrictHandlerFunc, operation string) openapi.StrictHandlerFunc {
	if !strings.HasPrefix(operation, "Identity") {
		return next
	}
	return func(ctx context.Context, w stdhttp.ResponseWriter, r *stdhttp.Request, request any) (any, error) {
		w.Header().Set("Cache-Control", "private, no-store")
		if r.Method != stdhttp.MethodGet {
			if err := app.ValidateIdentityOrigin(r.Header.Get("Origin")); err != nil {
				return nil, err
			}
			mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || mediaType != "application/json" {
				return nil, errx.NewInvalidInputError("identity requests must use application/json")
			}
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		peer, peerErr := netip.ParseAddr(ip)
		forwarded, forwardedErr := netip.ParseAddr(r.Header.Get("X-Real-IP"))
		if peerErr == nil && forwardedErr == nil && len(r.Header.Values("X-Real-IP")) == 1 {
			for _, prefix := range s.identityTrustedProxies {
				if prefix.Contains(peer.Unmap()) {
					ip = forwarded.Unmap().String()
					break
				}
			}
		}
		metadata := identityRequestMetadata{
			cookies: app.IdentityCookieHeader(strings.Join(r.Header.Values("Cookie"), "; ")),
			ip:      ip,
		}
		return next(context.WithValue(ctx, identityRequestKey{}, metadata), w, r, request)
	}
}

func (s *server) IdentityRegistrationCreate(
	ctx context.Context,
	request openapi.IdentityRegistrationCreateRequestObject,
) (openapi.IdentityRegistrationCreateResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	command := app.IdentityRegistration{IP: metadata.ip}
	if request.Body != nil {
		command.Email = string(request.Body.Email)
		command.DisplayName, command.Password = request.Body.DisplayName, request.Body.Password
	}
	result, err := s.application.IdentityRegister(ctx, command)
	if err != nil {
		s.logOperationError(ctx, "register identity account", err)
		return nil, err
	}
	cookie := identityCookie(result.Cookie)
	return openapi.IdentityRegistrationCreate201JSONResponse{
		Body:    identitySession(result.Session),
		Headers: openapi.IdentityRegistrationCreate201ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (s *server) IdentityLoginCreate(
	ctx context.Context,
	request openapi.IdentityLoginCreateRequestObject,
) (openapi.IdentityLoginCreateResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	command := app.IdentityLoginCommand{IP: metadata.ip}
	if request.Body != nil {
		command.Email, command.Password = string(request.Body.Email), request.Body.Password
	}
	result, err := s.application.IdentityLogin(ctx, command)
	if err != nil {
		s.logOperationError(ctx, "login identity account", err)
		return nil, err
	}
	cookie := identityCookie(result.Cookie)
	return openapi.IdentityLoginCreate200JSONResponse{
		Body:    identitySession(result.Session),
		Headers: openapi.IdentityLoginCreate200ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (s *server) IdentityLogout(
	ctx context.Context,
	request openapi.IdentityLogoutRequestObject,
) (openapi.IdentityLogoutResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	result, err := s.application.IdentityLogout(ctx, metadata.cookies)
	if err != nil {
		s.logOperationError(ctx, "logout identity account", err)
		return nil, err
	}
	cookie := identityCookie(result)
	return openapi.IdentityLogout204Response{
		Headers: openapi.IdentityLogout204ResponseHeaders{SetCookie: &cookie},
	}, nil
}

func (s *server) IdentitySessionGet(
	ctx context.Context,
	request openapi.IdentitySessionGetRequestObject,
) (openapi.IdentitySessionGetResponseObject, error) {
	metadata := ctx.Value(identityRequestKey{}).(identityRequestMetadata)
	result, err := s.application.IdentitySession(ctx, metadata.cookies)
	if err != nil {
		s.logOperationError(ctx, "read identity session", err)
		return nil, err
	}
	return openapi.IdentitySessionGet200JSONResponse(identitySession(result)), nil
}

func identityCookie(cookie app.IdentityCookie) string {
	return (&stdhttp.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Domain:   cookie.Domain,
		Path:     "/",
		Expires:  cookie.Expires,
		MaxAge:   cookie.MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: stdhttp.SameSiteLaxMode,
	}).String()
}

func identitySession(result app.IdentitySession) openapi.IdentitySession {
	account := nullable.NewNullNullable[openapi.IdentityAccount]()
	if session, ok := result.(app.IdentityAuthenticatedSession); ok {
		account = nullable.NewNullableWithValue(openapi.IdentityAccount{
			Id:              session.Account.ID,
			Email:           accountEmail(session.Account.Email.String()),
			DisplayName:     session.Account.DisplayName.String(),
			CreatedAt:       session.Account.CreatedAt,
			AuthenticatedAt: session.AuthenticatedAt,
		})
	}
	return openapi.IdentitySession{AccountApp: openapi.V2, Account: account}
}

// Responses keep the address part net/mail reads from a stored address and send an address it cannot read unchanged.
func accountEmail(stored string) string {
	address, err := mail.ParseAddress(stored)
	if err != nil {
		return stored
	}
	return address.Address
}

func (s *server) IdentityProxySessionCheck(
	ctx context.Context,
	request callbackopenapi.IdentityProxySessionCheckRequestObject,
) (callbackopenapi.IdentityProxySessionCheckResponseObject, error) {
	branch, cookies := "", app.IdentityCookieHeader("")
	if request.Params.XDevBranch != nil {
		branch = *request.Params.XDevBranch
	}
	if request.Params.Cookie != nil {
		cookies = app.IdentityCookieHeader(*request.Params.Cookie)
	}
	result, err := s.application.IdentityCheckSession(ctx, branch, cookies)
	if errors.Is(err, app.ErrIdentityTenantNotServed) {
		return callbackopenapi.IdentityProxySessionCheck421JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		s.logOperationError(ctx, "check identity proxy session", err)
		return nil, err
	}
	var response callbackopenapi.IdentityProxySession
	if session, ok := result.(app.IdentityAuthenticatedSession); ok {
		user := callbackopenapi.IdentityProxyUser{
			AuthenticatedAt: session.AuthenticatedAt,
			ExpiresAt:       session.ExpiresAt,
		}
		user.Identity.Id = session.Account.ID
		user.Identity.Traits = callbackopenapi.IdentityProxyTraits{
			Email:       accountEmail(session.Account.Email.String()),
			DisplayName: session.Account.DisplayName.String(),
		}
		err = response.FromIdentityProxyUser(user)
	} else {
		guest := callbackopenapi.IdentityProxyGuest{}
		guest.Identity.Id = "guest"
		err = response.FromIdentityProxyGuest(guest)
	}
	if err != nil {
		return nil, err
	}
	return callbackopenapi.IdentityProxySessionCheck200JSONResponse(response), nil
}

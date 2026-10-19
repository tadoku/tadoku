package app

import "github.com/tadoku/tadoku/services/tadoku-api/features/identity"

type IdentityRegistration = identity.Registration
type IdentityLoginCommand = identity.Login
type IdentityCookie = identity.Cookie
type IdentityCookieHeader = identity.CookieHeader
type IdentitySessionResult = identity.SessionResult
type IdentityAuthenticatedSession = identity.AuthenticatedSession
type IdentityError = identity.Error

const IdentityV2 = identity.V2

func ValidateIdentityOrigin(raw string) error   { return identity.ValidateOrigin(raw) }
func ValidateIdentityReturnTo(raw string) error { return identity.ValidateReturnTo(raw) }

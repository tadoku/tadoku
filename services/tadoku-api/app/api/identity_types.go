package api

import "github.com/tadoku/tadoku/services/tadoku-api/features/identity"

type IdentityRecovery = identity.Recovery
type IdentityRecoveryToken = identity.RecoveryToken
type IdentityProfileSettings = identity.ProfileSettings

type IdentityRegistration = identity.Registration
type IdentityLoginCommand = identity.Login
type IdentityCookie = identity.Cookie
type IdentityCookieHeader = identity.CookieHeader
type IdentitySession = identity.Session
type IdentityAuthenticatedSession = identity.AuthenticatedSession
type IdentityError = identity.Error

func ValidateIdentityOrigin(raw string) error   { return identity.ValidateOrigin(raw) }
func ValidateIdentityReturnTo(raw string) error { return identity.ValidateReturnTo(raw) }

package e2e_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type castIdentity struct {
	subject     string
	displayName string
	email       string
}

// Test safety: subjects and traits must match testdata/journeys/cast.json and testdata/tenant_accounts.sql.
var scenarioCast = map[member]castIdentity{
	guest:  {subject: "guest"},
	user:   {subject: "11111111-1111-4111-8111-111111111111", displayName: "Reader One", email: "reader@example.test"},
	admin:  {subject: "22222222-2222-4222-8222-222222222222", displayName: "Admin One", email: "admin@example.test"},
	banned: {subject: "33333333-3333-4333-8333-333333333333", displayName: "Banned One", email: "banned@example.test"},
	user2:  {subject: "44444444-4444-4444-8444-444444444444", displayName: "User Two", email: "user2@example.test"},
}

// Test safety: the key is generated per test process and never written anywhere; only this process's local JWKS
// publishes its public half.
var scenarioSigner struct {
	key *rsa.PrivateKey
	kid string
}

type castClaims struct {
	jwt.RegisteredClaims
	Type    string       `json:"type"`
	Session *castSession `json:"session"`
	Tenant  string       `json:"tenant"`
}

type castSession struct {
	Identity struct {
		Traits struct {
			DisplayName string `json:"display_name"`
			Email       string `json:"email"`
		} `json:"traits"`
	} `json:"identity"`
}

// fixtureJWKS returns the checked-in fixture JWKS plus the public key that signs scenario cast tokens.
func fixtureJWKS() ([]byte, error) {
	contents, err := os.ReadFile("testdata/authentication.jwks.json")
	if err != nil {
		return nil, err
	}
	var jwks struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(contents, &jwks); err != nil {
		return nil, fmt.Errorf("decode fixture JWKS: %w", err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate scenario signing key: %w", err)
	}
	scenarioSigner.key = key
	scenarioSigner.kid = "tadoku-scenario-" + uuid.NewString()
	jwks.Keys = append(jwks.Keys, map[string]string{
		"kty": "RSA",
		"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		"kid": scenarioSigner.kid,
		"use": "sig",
		"alg": "RS256",
	})
	return json.Marshal(jwks)
}

// signCastToken issues the checked-in cast's claims for key, valid one minute either side of the JWT verification
// clock.
func signCastToken(key tenant.Key, as member) (string, error) {
	claims, err := castClaimsFor(key.String(), as)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = scenarioSigner.kid
	return token.SignedString(scenarioSigner.key)
}

func castClaimsFor(tenantKey string, as member) (castClaims, error) {
	identity, ok := scenarioCast[as]
	if !ok {
		return castClaims{}, fmt.Errorf("cast has no member %q", as)
	}

	now := jwt.TimeFunc()
	claims := castClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://oathkeeper-api/",
			Subject:   identity.subject,
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
			NotBefore: jwt.NewNumericDate(now.Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		},
		Type:   "user",
		Tenant: tenantKey,
	}
	if as != guest {
		claims.Session = &castSession{}
		claims.Session.Identity.Traits.DisplayName = identity.displayName
		claims.Session.Identity.Traits.Email = identity.email
	}
	return claims, nil
}

// parseCastToken reads a checked-in fixture token's tenant and cast member without verifying its signature. It fails
// unless the claims are exactly the cast member's claims for that tenant, so re-signing them for another tenant
// changes nothing else.
func parseCastToken(token string) (string, member, error) {
	got := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(token, got); err != nil {
		return "", "", fmt.Errorf("parse fixture token: %w", err)
	}
	subject, _ := got["sub"].(string)
	tenantKey, _ := got["tenant"].(string)

	for as, identity := range scenarioCast {
		if identity.subject != subject {
			continue
		}
		claims, err := castClaimsFor(tenantKey, as)
		if err != nil {
			return "", "", err
		}
		encoded, err := json.Marshal(claims)
		if err != nil {
			return "", "", err
		}
		want := jwt.MapClaims{}
		if err := json.Unmarshal(encoded, &want); err != nil {
			return "", "", err
		}
		if !reflect.DeepEqual(got, want) {
			return "", "", fmt.Errorf("fixture token for %s in %q carries claims other than the cast's", as, tenantKey)
		}
		return tenantKey, as, nil
	}
	return "", "", fmt.Errorf("fixture token subject %q is not a cast member", subject)
}

package e2e_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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

var fixtureKeys map[string]*rsa.PublicKey

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
	fixtureKeys = make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		key, err := parseRSAPublicJWK(jwk)
		if err != nil {
			return nil, fmt.Errorf("fixture JWK %q: %w", jwk["kid"], err)
		}
		fixtureKeys[jwk["kid"]] = key
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

func parseRSAPublicJWK(jwk map[string]string) (*rsa.PublicKey, error) {
	n, err := base64.RawURLEncoding.DecodeString(jwk["n"])
	if err != nil {
		return nil, err
	}
	e, err := base64.RawURLEncoding.DecodeString(jwk["e"])
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}, nil
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

// resignFixtureToken re-signs a token that verifies against a checked-in fixture key with the scenario signing key,
// replacing only its tenant claim with stand(claim). Expired, future and other rejected claims stay as they are. A
// token that does not verify, or whose tenant claim is missing or not a tenant key, is returned unchanged, so a bad
// credential under test is never fixed.
func resignFixtureToken(token string, stand func(fixture string) string) (string, error) {
	claims := jwt.MapClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithJSONNumber(),
		jwt.WithoutClaimsValidation(),
	)
	_, err := parser.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		key, ok := fixtureKeys[kid]
		if !ok {
			return nil, fmt.Errorf("no checked-in fixture key %q", kid)
		}
		return key, nil
	})
	if err != nil {
		return token, nil
	}

	fixture, _ := claims["tenant"].(string)
	if _, err := tenant.Parse(fixture); err != nil {
		return token, nil
	}
	claims["tenant"] = stand(fixture)

	resigned := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	resigned.Header["kid"] = scenarioSigner.kid
	return resigned.SignedString(scenarioSigner.key)
}

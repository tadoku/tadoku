package serviceauth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// TokenValidator validates JWT tokens for service-to-service authentication
type TokenValidator struct {
	serviceName string
	publicKeys  map[string]*ecdsa.PublicKey // caller service name → public key
}

// NewTokenValidator creates a new token validator for the receiving service
// publicKeyDir should contain files named {service-name}.pub
func NewTokenValidator(serviceName string, publicKeyDir string) (*TokenValidator, error) {
	publicKeys := make(map[string]*ecdsa.PublicKey)

	entries, err := os.ReadDir(publicKeyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		callerName := strings.TrimSuffix(entry.Name(), ".pub")
		keyPath := filepath.Join(publicKeyDir, entry.Name())

		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key for %s: %w", callerName, err)
		}

		publicKey, err := parseECPublicKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key for %s: %w", callerName, err)
		}

		publicKeys[callerName] = publicKey
	}

	if len(publicKeys) == 0 {
		return nil, fmt.Errorf("no public keys found in %s", publicKeyDir)
	}

	return &TokenValidator{
		serviceName: serviceName,
		publicKeys:  publicKeys,
	}, nil
}

// NewTokenValidatorWithKeys creates a new token validator with pre-loaded keys
func NewTokenValidatorWithKeys(serviceName string, publicKeys map[string]*ecdsa.PublicKey) *TokenValidator {
	return &TokenValidator{
		serviceName: serviceName,
		publicKeys:  publicKeys,
	}
}

// Validate validates a service token and returns the calling service name
func (v *TokenValidator) Validate(tokenString string) (callingService string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (any, error) {
		// Verify signing method is exactly ES256
		if token.Method != jwt.SigningMethodES256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get claims to find issuer
		claims, ok := token.Claims.(*ServiceClaims)
		if !ok {
			return nil, fmt.Errorf("invalid claims type")
		}

		// Look up public key by issuer
		publicKey, exists := v.publicKeys[claims.Issuer]
		if !exists {
			return nil, fmt.Errorf("unknown issuer: %s", claims.Issuer)
		}

		return publicKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*ServiceClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	// Verify subject matches issuer (per RFC 7523 for service tokens)
	if claims.Subject != claims.Issuer {
		return "", fmt.Errorf("invalid service token: subject (%s) does not match issuer (%s)", claims.Subject, claims.Issuer)
	}

	// Verify audience matches this service
	if !claims.VerifyAudience(v.serviceName, true) {
		return "", fmt.Errorf("token not intended for this service (audience: %v)", claims.Audience)
	}

	return claims.Issuer, nil
}

// parseECPublicKey parses a PEM-encoded ECDSA public key
func parseECPublicKey(pemData []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as PKIX public key (most common format)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not an ECDSA public key")
	}

	return ecPub, nil
}

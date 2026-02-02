package serviceauth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TokenGenerator creates JWT tokens for service-to-service authentication
type TokenGenerator struct {
	serviceName string
	privateKey  *ecdsa.PrivateKey
	clock       func() time.Time
	expiry      time.Duration
}

// NewTokenGenerator creates a new token generator for the given service
func NewTokenGenerator(serviceName string, privateKeyPEM []byte) (*TokenGenerator, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	privateKey, err := parseECPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &TokenGenerator{
		serviceName: serviceName,
		privateKey:  privateKey,
		clock:       time.Now,
		expiry:      TokenExpiry,
	}, nil
}

// WithExpiry sets a custom token expiry duration
func (g *TokenGenerator) WithExpiry(expiry time.Duration) *TokenGenerator {
	g.expiry = expiry
	return g
}

// NewTokenGeneratorFromFile creates a new token generator loading the key from a file
func NewTokenGeneratorFromFile(serviceName string, privateKeyPath string) (*TokenGenerator, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return NewTokenGenerator(serviceName, keyData)
}

// Generate creates a signed JWT token for the target service
func (g *TokenGenerator) Generate(targetService string) (string, error) {
	if targetService == "" {
		return "", fmt.Errorf("target service cannot be empty")
	}

	claims := NewServiceClaimsWithExpiry(g.serviceName, targetService, g.clock(), g.expiry)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signedToken, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ServiceName returns the name of the service this generator creates tokens for
func (g *TokenGenerator) ServiceName() string {
	return g.serviceName
}

// parseECPrivateKey parses a PEM-encoded ECDSA private key
func parseECPrivateKey(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as PKCS8 first (more common format)
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if ecKey, ok := key.(*ecdsa.PrivateKey); ok {
			return ecKey, nil
		}
		return nil, fmt.Errorf("key is not an ECDSA private key")
	}

	// Fall back to EC private key format
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	return key, nil
}

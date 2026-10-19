package tokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	repository *Repository
	production bool
}

var (
	_ Verifier = (*Service)(nil)
	_ Manager  = (*Service)(nil)
)

func NewService(repository *Repository, production bool) *Service {
	return &Service{repository: repository, production: production}
}

func (s *Service) Verify(ctx context.Context, raw string, deployment tenant.Deployment) (Principal, error) {
	key, err := tokenTenant(raw)
	if err != nil {
		return Principal{}, err
	}
	if !deployment.Serves(key) {
		return Principal{}, ErrTenantNotServed
	}
	if s.production && key == tenant.Production() {
		return Principal{}, errx.NewUnauthorizedError("invalid or expired access token")
	}
	if current, ok := tenant.FromContext(ctx); ok && current != key {
		return Principal{}, errx.NewUnauthorizedError("access token tenant mismatch")
	}
	row, err := s.repository.verify(tenant.WithKey(ctx, key), tokenHash(raw), timex.Now())
	if err != nil {
		return Principal{}, err
	}
	return Principal{Tenant: key, TokenID: row.ID, Scopes: row.Scopes}, nil
}

// Create requires the caller's tenant transaction; publish Secret only after commit, through an attached invocation.
func (s *Service) Create(ctx context.Context, request Create) (Issued, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return Issued{}, postgres.ErrNoTenant
	}
	if s.production && key == tenant.Production() {
		return Issued{}, errx.NewForbiddenError("access tokens are unavailable for tadoku/prod in production")
	}
	if strings.TrimSpace(request.Name) == "" || len(request.Name) > 128 {
		return Issued{}, errx.NewInvalidInputError("token name must contain 1–128 bytes")
	}
	if len(request.Scopes) != 1 || request.Scopes[0] != EmailRead {
		return Issued{}, errx.NewInvalidInputError("token scope must be email:read")
	}
	if request.ExpiresIn <= 0 {
		return Issued{}, errx.NewInvalidInputError("token expiry must be positive")
	}
	if request.ExpiresIn > 30*24*time.Hour {
		return Issued{}, errx.NewInvalidInputError("token expiry must not exceed 30 days")
	}
	if err := s.repository.lockTenant(ctx); err != nil {
		return Issued{}, err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return Issued{}, err
	}
	raw := "tdk_at_" + strings.ReplaceAll(key.String(), "/", "_") + "_" + base64.RawURLEncoding.EncodeToString(secret)
	now := timex.Now()
	row := Token{
		ID:        uuid.New(),
		Name:      request.Name,
		Scopes:    []Scope{EmailRead},
		CreatedAt: now,
		ExpiresAt: now.Add(request.ExpiresIn),
	}
	if err := s.repository.insert(ctx, row, tokenHash(raw)); err != nil {
		return Issued{}, err
	}
	return Issued{Token: row, Secret: raw}, nil
}

func (s *Service) List(ctx context.Context) ([]Token, error) { return s.repository.list(ctx) }
func (s *Service) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.repository.revoke(ctx, id, timex.Now())
}

func tokenHash(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func tokenTenant(raw string) (tenant.Key, error) {
	invalid := errx.NewUnauthorizedError("invalid or expired access token")
	rest, ok := strings.CutPrefix(raw, "tdk_at_")
	if !ok {
		return tenant.Key{}, invalid
	}
	name, rest, ok := strings.Cut(rest, "_")
	if !ok {
		return tenant.Key{}, invalid
	}
	id, secret, ok := strings.Cut(rest, "_")
	if !ok {
		return tenant.Key{}, invalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil || len(decoded) != 32 {
		return tenant.Key{}, invalid
	}
	key, err := tenant.Parse(name + "/" + id)
	if err != nil {
		return tenant.Key{}, invalid
	}
	return key, nil
}

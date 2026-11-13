package tokens

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestCreateRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, false).Create(t.Context(), Create{})
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("Create without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestListRequiresTenant(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, false).List(t.Context())
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("List without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestRevokeRequiresTenant(t *testing.T) {
	t.Parallel()

	err := NewService(nil, false).Revoke(t.Context(), uuid.Nil)
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, postgres.ErrNoTenant) {
		t.Errorf("Revoke without tenant = %v, want unavailable wrapping %v", err, postgres.ErrNoTenant)
	}
}

func TestCreateTrimsNameBeforeValidateAndPersist(t *testing.T) {
	t.Parallel()

	ctx := tenant.WithKey(t.Context(), tenant.Production())
	repo := &createRepositoryStub{}
	service := &Service{repository: repo, production: false}
	request := Create{Scopes: []Scope{EmailRead}, ExpiresIn: time.Hour}

	request.Name = "  ops  "
	issued, err := service.Create(ctx, request)
	if err != nil {
		t.Fatalf("padded name create: %v", err)
	}
	if issued.Token.Name != "ops" || repo.insertedName != "ops" {
		t.Errorf("padded name stored %q (insert %q), want %q", issued.Token.Name, repo.insertedName, "ops")
	}

	request.Name = strings.Repeat(" ", 60) + "ops" + strings.Repeat(" ", 70)
	if len(request.Name) <= 128 {
		t.Fatalf("heavyPad len = %d, want > 128 raw bytes", len(request.Name))
	}
	issued, err = service.Create(ctx, request)
	if err != nil {
		t.Fatalf("heavy padding create: %v", err)
	}
	if issued.Token.Name != "ops" || repo.insertedName != "ops" {
		t.Errorf("heavy padding stored %q (insert %q), want %q", issued.Token.Name, repo.insertedName, "ops")
	}

	insertsBefore := repo.inserts
	request.Name = "   \t  "
	_, err = service.Create(ctx, request)
	if errx.KindOf(err) != errx.InvalidInput || err.Error() != "token name must contain 1–128 bytes" {
		t.Errorf("whitespace-only name = %v, want invalid input token name must contain 1–128 bytes", err)
	}
	if repo.inserts != insertsBefore {
		t.Errorf("whitespace-only name inserted %d times, want no insert", repo.inserts-insertsBefore)
	}
}

type createRepositoryStub struct {
	insertedName string
	inserts      int
}

func (createRepositoryStub) lockTenant(context.Context) error { return nil }

func (s *createRepositoryStub) insert(_ context.Context, row Token, _ string, _ int) (bool, error) {
	s.inserts++
	s.insertedName = row.Name
	return true, nil
}

func (createRepositoryStub) verify(context.Context, string, []Scope, time.Time) (Token, bool, error) {
	return Token{}, false, errors.New("unexpected verify")
}

func (createRepositoryStub) list(context.Context) ([]Token, error) {
	return nil, errors.New("unexpected list")
}

func (createRepositoryStub) revoke(context.Context, uuid.UUID, time.Time) error {
	return errors.New("unexpected revoke")
}

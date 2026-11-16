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

func TestCreateStoresValidNamesUnchanged(t *testing.T) {
	t.Parallel()

	names := []string{
		"ops",
		"email-debug_2026",
		"a/b=c+d[e](f){g}",
		strings.Repeat("x", 128),
	}
	for _, name := range names {
		repo := &createRepositoryStub{}
		issued, err := createTokenNamed(t, repo, name)
		if err != nil {
			t.Fatalf("create %q: %v", name, err)
		}
		if issued.Token.Name != name || repo.insertedName != name {
			t.Errorf("create %q stored %q (insert %q), want unchanged", name, issued.Token.Name, repo.insertedName)
		}
	}
}

func TestCreateRejectsInvalidNamesWithoutInsert(t *testing.T) {
	t.Parallel()

	names := []string{
		"",
		"email debug",
		"  ops  ",
		"ops\t",
		"ops\n",
		"ops!",
		"ops.prod",
		"ops@tadoku",
		"été",
		strings.Repeat("x", 129),
	}
	for _, name := range names {
		repo := &createRepositoryStub{}
		_, err := createTokenNamed(t, repo, name)
		if errx.KindOf(err) != errx.InvalidInput {
			t.Errorf("create %q = %v, want invalid input", name, err)
		}
		if repo.inserts != 0 {
			t.Errorf("create %q inserted %d times, want no insert", name, repo.inserts)
		}
	}
}

func TestCreateRejectsActiveTokenCapAsConflict(t *testing.T) {
	t.Parallel()

	repo := &createRepositoryStub{rejectInsert: true}
	_, err := createTokenNamed(t, repo, "ops")
	if errx.KindOf(err) != errx.Conflict {
		t.Errorf("active-token cap = %v, want conflict", err)
	}
	if err == nil || err.Error() != "tenant already has 10 active tokens" {
		t.Errorf("active-token cap message = %v, want tenant already has 10 active tokens", err)
	}
	if repo.inserts != 1 {
		t.Errorf("active-token cap inserted %d times, want 1", repo.inserts)
	}
}

func createTokenNamed(t *testing.T, repo *createRepositoryStub, name string) (Issued, error) {
	t.Helper()

	ctx := tenant.WithKey(t.Context(), tenant.Production())
	service := &Service{repository: repo, production: false}
	request := Create{Name: name, Scopes: []Scope{EmailRead}, ExpiresIn: time.Hour}

	return service.Create(ctx, request)
}

type createRepositoryStub struct {
	insertedName string
	inserts      int
	rejectInsert bool
}

func (createRepositoryStub) lockTenant(context.Context) error { return nil }

func (s *createRepositoryStub) insert(_ context.Context, row Token, _ string, _ int) (bool, error) {
	s.inserts++
	s.insertedName = row.Name
	if s.rejectInsert {
		return false, nil
	}
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

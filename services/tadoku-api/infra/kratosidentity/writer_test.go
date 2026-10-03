package kratosidentity_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	kratosclient "github.com/tadoku/tadoku/services/tadoku-api/infra/kratos"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/kratosidentity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testkratos"
)

func TestWriterTenantBoundaries(t *testing.T) {
	fixture, err := testkratos.New(t.Context(), "testdata/identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	identityID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testKey, err := tenant.Parse("e2e/kratos-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	contexts := []struct {
		name string
		ctx  context.Context
	}{
		{"production", tenant.WithKey(t.Context(), tenant.Production())},
		{"test tenant", tenant.WithKey(t.Context(), testKey)},
		{"no tenant", t.Context()},
		{"zero tenant", tenant.WithKey(t.Context(), tenant.Key{})},
	}

	var logs bytes.Buffer
	providerConfig := fixture.Client().GetConfig()
	writer := kratosidentity.NewWriter(
		providerConfig.Servers[0].URL,
		providerConfig.HTTPClient,
		slog.New(slog.NewTextHandler(&logs, nil)),
	)
	operations := []struct {
		name string
		call func(context.Context, uuid.UUID) (kratosidentity.Outcome, error)
	}{
		{"deactivate", writer.Deactivate},
		{"delete sessions", writer.DeleteSessions},
		{"delete", writer.Delete},
	}

	for _, operation := range operations {
		for _, scope := range contexts {
			t.Run(operation.name+"/"+scope.name, func(t *testing.T) {
				if err := fixture.Reset(t.Context()); err != nil {
					t.Fatal(err)
				}
				before, err := fixture.CursorClient().FetchIdentity(t.Context(), identityID)
				if err != nil || before.GetState() != "active" {
					t.Fatalf("active identity prerequisite: identity=%v error=%v", before, err)
				}
				sessions, _, err := fixture.Client().IdentityApi.ListIdentitySessions(
					t.Context(), identityID.String(),
				).Execute()
				if err != nil || len(sessions) != 1 || !sessions[0].GetActive() {
					t.Fatalf("active session prerequisite: sessions=%v error=%v", sessions, err)
				}
				logs.Reset()

				outcome, writeErr := operation.call(scope.ctx, identityID)
				after, readErr := fixture.CursorClient().FetchIdentity(t.Context(), identityID)
				afterSessions, _, sessionsErr := fixture.Client().IdentityApi.ListIdentitySessions(
					t.Context(), identityID.String(),
				).Execute()

				if scope.name == "production" {
					if writeErr != nil || outcome != kratosidentity.Applied {
						t.Fatalf("production write: outcome=%v error=%v", outcome, writeErr)
					}
					switch operation.name {
					case "deactivate":
						if readErr != nil || after.GetState() != "inactive" ||
							!reflect.DeepEqual(after.GetTraits(), before.GetTraits()) ||
							!reflect.DeepEqual(after.GetMetadataAdmin(), before.GetMetadataAdmin()) ||
							!reflect.DeepEqual(after.GetMetadataPublic(), before.GetMetadataPublic()) {
							t.Errorf(
								"deactivation changed unrelated identity data: before=%v after=%v error=%v",
								before, after, readErr,
							)
						}
					case "delete sessions":
						if readErr != nil || !reflect.DeepEqual(after, before) ||
							sessionsErr != nil || len(afterSessions) != 0 {
							t.Errorf(
								"session deletion: identity=%v error=%v sessions=%v error=%v",
								after, readErr, afterSessions, sessionsErr,
							)
						}
					case "delete":
						if !errors.Is(readErr, kratosclient.ErrNotFound) {
							t.Errorf("deleted identity read: identity=%v error=%v", after, readErr)
						}
					}
					repeated, err := operation.call(scope.ctx, identityID)
					if err != nil || repeated != kratosidentity.Applied {
						t.Errorf("repeated write: outcome=%v error=%v", repeated, err)
					}
					return
				}

				if scope.name == "test tenant" {
					if writeErr != nil || outcome != kratosidentity.SkippedForTestTenant {
						t.Errorf("test tenant write: outcome=%v error=%v", outcome, writeErr)
					}
					if !strings.Contains(logs.String(), "tenant="+testKey.String()) ||
						!strings.Contains(logs.String(), "identity_id="+identityID.String()) {
						t.Errorf("skip log missing tenant or identity: %s", logs.String())
					}
				} else if writeErr == nil {
					t.Error("missing tenant accepted identity write")
				}
				if readErr != nil || !reflect.DeepEqual(after, before) ||
					sessionsErr != nil || !reflect.DeepEqual(afterSessions, sessions) {
					t.Errorf(
						"guard changed identity or sessions: identity=%v error=%v sessions=%v error=%v",
						after, readErr, afterSessions, sessionsErr,
					)
				}
			})
		}
	}
}

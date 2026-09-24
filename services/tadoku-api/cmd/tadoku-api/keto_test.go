package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

func TestApplicationKetoRelationships(t *testing.T) {
	fixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	cfg := validApplicationConfig(t)
	cfg.KetoReadURL = fixture.ReadURL()
	cfg.KetoWriteURL = fixture.WriteURL()
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	application, err := start(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		if err := application.wait(); err != nil {
			t.Error(err)
		}
	})
	client := application.keto

	for _, encoding := range []string{"direct", "subject set"} {
		t.Run(encoding, func(t *testing.T) {
			if err := fixture.Reset(t.Context()); err != nil {
				t.Fatal(err)
			}
			subject := ketoclient.Subject{ID: "user /?&+"}
			otherSubjects := []ketoclient.Subject{{ID: "other-user"}}
			if encoding == "subject set" {
				subject = ketoclient.Subject{Set: &ketoclient.SubjectSet{
					Namespace: "app",
					Object:    "group /?&+",
					Relation:  "admins",
				}}
				otherSubjects = []ketoclient.Subject{
					{Set: &ketoclient.SubjectSet{Namespace: "User", Object: "group /?&+", Relation: "admins"}},
					{Set: &ketoclient.SubjectSet{Namespace: "app", Object: "other-group", Relation: "admins"}},
					{Set: &ketoclient.SubjectSet{Namespace: "app", Object: "group /?&+", Relation: "banned"}},
					{ID: "group /?&+"},
				}
			}
			target := ketoclient.PermissionCheck{
				Namespace: "app",
				Object:    "resource /?&+",
				Relation:  "admins",
				Subject:   subject,
			}
			// Each neighbor differs in exactly one tuple component.
			tuples := []ketoclient.PermissionCheck{target, target, target, target}
			tuples[1].Namespace = "User"
			tuples[2].Object = "other-resource"
			tuples[3].Relation = "banned"
			for _, other := range otherSubjects {
				neighbor := target
				neighbor.Subject = other
				tuples = append(tuples, neighbor)
			}
			for index, tuple := range tuples {
				for range 2 {
					if err := client.AddRelation(t.Context(), tuple.Namespace, tuple.Object, tuple.Relation, tuple.Subject); err != nil {
						t.Fatalf("add tuple %d: %v", index, err)
					}
				}
				allowed, err := client.CheckPermission(t.Context(), tuple.Namespace, tuple.Object, tuple.Relation, tuple.Subject)
				if err != nil || !allowed {
					t.Fatalf("check added tuple %d: allowed=%v error=%v", index, allowed, err)
				}
			}
			for range 2 {
				if err := client.DeleteRelation(t.Context(), target.Namespace, target.Object, target.Relation, target.Subject); err != nil {
					t.Fatal(err)
				}
				for index, tuple := range tuples {
					allowed, err := client.CheckPermission(t.Context(), tuple.Namespace, tuple.Object, tuple.Relation, tuple.Subject)
					if err != nil || allowed != (index != 0) {
						t.Errorf("check after exact delete, tuple %d: allowed=%v error=%v", index, allowed, err)
					}
				}
			}
		})
	}
}

func TestApplicationBoundsKetoWrites(t *testing.T) {
	cfg := validApplicationConfig(t)
	provider := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	t.Cleanup(provider.Close)
	cfg.KetoWriteURL = provider.URL
	cfg.KetoWriteTimeout = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	application, err := start(ctx, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("startup without responsive Keto write provider: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		if err := application.wait(); err != nil {
			t.Error(err)
		}
	})

	for _, operation := range []string{"add", "delete"} {
		t.Run(operation, func(t *testing.T) {
			write := application.keto.AddRelation
			if operation == "delete" {
				write = application.keto.DeleteRelation
			}
			ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer cancel()
			err := write(ctx, "app", "synthetic", "admins", ketoclient.Subject{ID: "synthetic"})
			if !errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
				t.Errorf("Keto write error=%v, caller error=%v", err, ctx.Err())
			}
		})
	}
}

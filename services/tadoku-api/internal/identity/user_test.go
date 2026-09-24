package identity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestActorID(t *testing.T) {
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	tests := []struct {
		name string
		user *identity.User
		want uuid.UUID
		ok   bool
	}{
		{name: "missing identity"},
		{name: "guest", user: &identity.User{Subject: "guest"}},
		{name: "nil UUID", user: &identity.User{Subject: uuid.Nil.String()}},
		{name: "user", user: &identity.User{Subject: userID.String()}, want: userID, ok: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.user != nil {
				ctx = identity.WithUser(ctx, test.user)
			}

			got, ok := identity.ActorID(ctx)
			if got != test.want || ok != test.ok {
				t.Errorf("ActorID = %v, %v; want %v, %v", got, ok, test.want, test.ok)
			}

			got, err := identity.RequireActorID(ctx)
			if test.ok {
				if err != nil || got != test.want {
					t.Errorf("RequireActorID = %v, %v; want %v, nil", got, err, test.want)
				}
			} else if errx.KindOf(err) != errx.Unauthorized {
				t.Errorf("RequireActorID error = %v, want unauthorized", err)
			}
		})
	}
}

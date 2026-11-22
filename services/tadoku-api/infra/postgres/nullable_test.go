package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func TestUUIDsSkipsInvalidElements(t *testing.T) {
	t.Parallel()

	valid := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	values := []pgtype.UUID{
		{Bytes: valid, Valid: true},
		{},
		{Bytes: uuid.Nil, Valid: false},
	}

	got := postgres.UUIDs(values)

	if len(got) != 1 {
		t.Fatalf("UUIDs() len=%d, want 1", len(got))
	}
	if got[0] != valid {
		t.Errorf("UUIDs()[0]=%v, want %v", got[0], valid)
	}
	for _, id := range got {
		if id == uuid.Nil {
			t.Errorf("UUIDs() returned uuid.Nil for a NULL element")
		}
	}
}

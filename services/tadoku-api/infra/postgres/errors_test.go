package postgres_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "wrapped unique violation", err: fmt.Errorf("create: %w", &pgconn.PgError{Code: pgerrcode.UniqueViolation}), want: true},
		{name: "other postgres error", err: &pgconn.PgError{Code: pgerrcode.CheckViolation}, want: false},
		{name: "plain error", err: errors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := postgres.IsUniqueViolation(test.err); got != test.want {
				t.Errorf("IsUniqueViolation()=%v, want %v", got, test.want)
			}
		})
	}
}

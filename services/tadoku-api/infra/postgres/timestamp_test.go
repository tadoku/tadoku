package postgres_test

import (
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func TestTimestampConvertsToUTC(t *testing.T) {
	location := time.FixedZone("UTC+9", 9*60*60)
	value := time.Date(2026, 9, 13, 10, 0, 0, 123456789, location)

	got := postgres.Timestamp(value)
	if !got.Valid {
		t.Fatal("timestamp is invalid")
	}
	if !got.Time.Equal(value) {
		t.Errorf("timestamp instant=%v, want %v", got.Time, value)
	}
	if got.Time.Location() != time.UTC {
		t.Errorf("timestamp location=%v, want UTC", got.Time.Location())
	}
}

package uuidx

import (
	"context"
	"io"
	"sync"

	"github.com/google/uuid"
)

type source struct {
	mu     sync.Mutex
	reader io.Reader
}

type sourceKey struct{}

// WithSource makes New read random bytes from reader for ctx and every context derived from it, in call order.
// Production code never attaches one.
func WithSource(ctx context.Context, reader io.Reader) context.Context {
	return context.WithValue(ctx, sourceKey{}, &source{reader: reader})
}

// New returns a random version 4 UUID for a new record. It panics when the attached source fails, as uuid.New does.
func New(ctx context.Context) uuid.UUID {
	s, ok := ctx.Value(sourceKey{}).(*source)
	if !ok {
		return uuid.New()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return uuid.Must(uuid.NewRandomFromReader(s.reader))
}

// NewV7 returns a time-ordered version 7 UUID, reading the same attached source as New. It panics when that source
// fails.
func NewV7(ctx context.Context) uuid.UUID {
	s, ok := ctx.Value(sourceKey{}).(*source)
	if !ok {
		return uuid.Must(uuid.NewV7())
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return uuid.Must(uuid.NewV7FromReader(s.reader))
}

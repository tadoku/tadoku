package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
)

type Policy struct {
	Concurrency int
	Timeout     time.Duration
	MaxAttempts int
}

type registration struct {
	spec   policy
	invoke func(context.Context, json.RawMessage) error
	err    error
}

type registry struct {
	ordered []registration
	byType  map[jobs.Type]registration
}

func handle[J jobs.Job](fn func(context.Context, J) error, spec Policy) registration {
	if reflect.TypeFor[J]().Kind() != reflect.Struct {
		return registration{err: errors.New("job handler requires a concrete value payload")}
	}
	if fn == nil {
		return registration{err: errors.New("job handler is nil")}
	}
	var job J
	name := job.Type()
	if name == "" || spec.Concurrency < 1 || spec.Concurrency > 100 || spec.Timeout <= 0 || spec.MaxAttempts < 1 || spec.MaxAttempts > math.MaxInt32 {
		return registration{err: fmt.Errorf("invalid handler policy for %q", name)}
	}
	return registration{
		spec: policy{typeName: name, limit: spec.Concurrency, timeout: spec.Timeout, lease: 10 * time.Second, maxAttempts: spec.MaxAttempts},
		invoke: func(ctx context.Context, raw json.RawMessage) error {
			var payload J
			if err := decode(raw, &payload); err != nil {
				return &PermanentError{Err: err}
			}
			if err := payload.Validate(); err != nil {
				return &PermanentError{Err: err}
			}
			return fn(ctx, payload)
		},
	}
}

func newRegistry(entries ...registration) (*registry, error) {
	if len(entries) == 0 {
		return nil, errors.New("worker requires at least one registered handler")
	}
	handlers := &registry{ordered: append([]registration(nil), entries...), byType: make(map[jobs.Type]registration, len(entries))}
	for _, entry := range entries {
		if entry.err != nil {
			return nil, entry.err
		}
		if entry.invoke == nil {
			return nil, errors.New("job handler is nil")
		}
		if _, exists := handlers.byType[entry.spec.typeName]; exists {
			return nil, fmt.Errorf("duplicate job handler %q", entry.spec.typeName)
		}
		handlers.byType[entry.spec.typeName] = entry
	}
	return handlers, nil
}

func (r *registry) types() []jobs.Type {
	types := make([]jobs.Type, 0, len(r.ordered))
	for _, entry := range r.ordered {
		types = append(types, entry.spec.typeName)
	}
	return types
}

func (r *registry) dispatch(ctx context.Context, job jobqueue.ClaimedJob) error {
	entry, exists := r.byType[job.Type]
	if !exists {
		return &UnknownTypeError{Type: job.Type}
	}
	return entry.invoke(ctx, job.Payload)
}

func decode(raw json.RawMessage, dst any) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return errors.New("job payload must contain one object")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode job payload: %w", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("job payload must contain one object")
	}
	return nil
}

type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

type UnknownTypeError struct{ Type jobs.Type }

func (e *UnknownTypeError) Error() string { return fmt.Sprintf("unsupported job type %q", e.Type) }

func failureCode(err error) string {
	var permanent *PermanentError
	if errors.As(err, &permanent) {
		return "invalid_payload"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	return "handler_error"
}

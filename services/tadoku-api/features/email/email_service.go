package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"maps"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Sender interface {
	Send(context.Context, uuid.UUID, string, string, string) (string, error)
}

type Service struct {
	repository *Repository
	sender     Sender
}

var (
	_ Writer = (*Service)(nil)
	_ Reader = (*Service)(nil)
)

// A nil sender holds every email. Supply a sender only in deployments that enable delivery.
func NewService(repository *Repository, sender Sender) *Service {
	return &Service{repository: repository, sender: sender}
}

func (s *Service) Write(ctx context.Context, message mail.Message) (WriteResult, error) {
	subject, body, err := render(message)
	if err != nil {
		return WriteResult{}, err
	}

	row := Email{
		ID:        uuid.New(),
		Template:  message.Template,
		Recipient: message.Recipient,
		Subject:   subject,
		Body:      body,
		Params:    maps.Clone(message.Params),
		Delivery:  Pending{},
		CreatedAt: timex.Now(),
	}
	if err := s.repository.insert(ctx, row); err != nil {
		return WriteResult{}, err
	}

	return WriteResult{EmailID: row.ID, Jobs: []jobs.Job{jobs.SendEmailV1{EmailID: row.ID}}}, nil
}

type cursor struct {
	Tenant string    `json:"tenant"`
	At     time.Time `json:"at"`
	ID     uuid.UUID `json:"id"`
}

func (s *Service) List(ctx context.Context, filter Filter) (Page, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return Page{}, errx.NewUnavailableError("email tenant missing", nil)
	}
	if filter.Limit == 0 {
		filter.Limit = 100
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		return Page{}, errx.NewInvalidInputError("email limit must be 1 to 100")
	}
	switch filter.Status {
	case "", StatusPending, StatusSent, StatusHeld, StatusFailed:
	default:
		return Page{}, errx.NewInvalidInputError("invalid email status")
	}

	var before *cursor
	if filter.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(filter.Cursor)
		value := cursor{}
		if err != nil || len(filter.Cursor) > 2048 || json.Unmarshal(raw, &value) != nil ||
			value.Tenant != key.String() || value.ID == uuid.Nil || value.At.IsZero() {
			return Page{}, errx.NewInvalidInputError("invalid email cursor")
		}
		before = &value
	}

	rows, err := s.repository.list(ctx, filter, before)
	if err != nil {
		return Page{}, err
	}

	result := Page{Emails: rows}
	if len(rows) > filter.Limit {
		result.Emails = rows[:filter.Limit]
		last := result.Emails[len(result.Emails)-1]
		raw, err := json.Marshal(cursor{Tenant: key.String(), At: last.CreatedAt, ID: last.ID})
		if err != nil {
			return Page{}, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
	}
	return result, nil
}

func (s *Service) Find(ctx context.Context, id uuid.UUID) (Email, error) {
	return s.repository.find(ctx, id)
}

func (s *Service) Send(ctx context.Context, id uuid.UUID) error { return s.SendAttempt(ctx, id, false) }

// final is set only for the queue's last delivery attempt.
func (s *Service) SendAttempt(ctx context.Context, id uuid.UUID, final bool) error {
	row, err := s.Find(ctx, id)
	if errx.KindOf(err) == errx.NotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if _, ok := row.Delivery.(Pending); !ok {
		return nil
	}
	if row.Tenant != tenant.Production() || s.sender == nil {
		return s.repository.finish(ctx, id, Held{})
	}

	messageID, err := s.sender.Send(ctx, id, row.Recipient, row.Subject, row.Body)
	if err != nil {
		if final {

			finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			defer cancel()
			return errors.Join(err, s.repository.finish(finishCtx, id, Failed{Error: err.Error()}))
		}
		return err
	}

	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	return s.repository.finish(finishCtx, id, Sent{At: timex.Now(), ProviderMessageID: messageID})
}

func (s *Service) Retain(ctx context.Context) error {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return errx.NewUnavailableError("email tenant missing", nil)
	}

	retention := 14 * 24 * time.Hour
	if key == tenant.Production() {
		retention = 3 * 24 * time.Hour
	}
	return s.repository.retain(ctx, timex.Now().Add(-retention))
}

func (s *Service) ReconcileFailures(ctx context.Context) error {
	return s.repository.reconcileFailures(ctx, s.sender == nil)
}

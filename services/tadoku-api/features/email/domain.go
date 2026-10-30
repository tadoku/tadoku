package email

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusHeld    Status = "held"
	StatusFailed  Status = "failed"
)

type Delivery interface{ delivery() }
type Pending struct{}

func (Pending) delivery() {}

type Sent struct {
	At                time.Time
	ProviderMessageID string
}

func (Sent) delivery() {}

type Held struct{}

func (Held) delivery() {}

type Failed struct{ Error string }

func (Failed) delivery() {}

type Email struct {
	ID                       uuid.UUID
	Tenant                   tenant.Key
	Template                 mail.Template
	Recipient, Subject, Body string
	Params                   map[string]string
	Delivery                 Delivery
	CreatedAt                time.Time
}
type Filter struct {
	Recipient string
	Template  mail.Template
	Status    Status
	Since     *time.Time
	Limit     int
	Cursor    string
}
type Page struct {
	Emails     []Email
	NextCursor string
}
type SendFailure struct {
	EmailID uuid.UUID
	Error   string
}
type WriteResult struct {
	EmailID uuid.UUID
	Jobs    []jobs.Job
}

type Writer interface {
	// Write requires an enclosing business transaction; persist all returned jobs before committing.
	Write(context.Context, mail.Message) (WriteResult, error)
}
type Reader interface {
	List(context.Context, Filter) (Page, error)
	Find(context.Context, uuid.UUID) (Email, error)
}

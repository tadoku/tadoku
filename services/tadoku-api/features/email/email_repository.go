package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/email"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) insert(ctx context.Context, row Email) error {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	if _, ok := executor.(pgx.Tx); !ok {
		return errors.New("email write requires an enclosing business transaction")
	}
	params, err := json.Marshal(row.Params)
	if err != nil {
		return err
	}
	return queries.New(executor).Insert(ctx, queries.InsertParams{
		ID:        postgres.UUID(row.ID),
		Template:  string(row.Template),
		Recipient: row.Recipient,
		Subject:   row.Subject,
		Body:      row.Body,
		Params:    params,
		CreatedAt: postgres.Timestamptz(row.CreatedAt),
	})
}

func (r *Repository) find(ctx context.Context, id uuid.UUID) (Email, error) {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return Email{}, err
	}
	row, err := queries.New(executor).Find(ctx, postgres.UUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Email{}, errx.NewNotFoundError("email not found")
	}
	if err != nil {
		return Email{}, err
	}
	return emailFromRow(row)
}

func (r *Repository) list(ctx context.Context, filter Filter, before *cursor) ([]Email, error) {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	args := queries.ListParams{
		Recipient: filter.Recipient,
		Template:  string(filter.Template),
		Status:    string(filter.Status),
		PageSize:  int32(filter.Limit + 1),
	}
	if filter.Since != nil {
		args.Since = postgres.Timestamptz(*filter.Since)
	}
	if before != nil {
		args.BeforeTime = postgres.Timestamptz(before.At)
		args.BeforeID = postgres.UUID(before.ID)
	}
	rows, err := queries.New(executor).List(ctx, args)
	if err != nil {
		return nil, err
	}
	result := make([]Email, 0, len(rows))
	for _, row := range rows {
		item, err := emailFromRow(queries.FindRow(row))
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Repository) finish(ctx context.Context, id uuid.UUID, delivery Delivery) error {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	args := queries.FinishParams{ID: postgres.UUID(id)}
	switch value := delivery.(type) {
	case Held:
		args.Status = string(StatusHeld)
	case Sent:
		args.Status = string(StatusSent)
		args.SentAt = postgres.Timestamptz(value.At)
		args.ProviderMessageID = postgres.NullableText(&value.ProviderMessageID)
	case Failed:
		args.Status = string(StatusFailed)
		args.Error = postgres.NullableText(&value.Error)
	default:
		return fmt.Errorf("invalid email completion %T", delivery)
	}
	return queries.New(executor).Finish(ctx, args)
}

func (r *Repository) retain(ctx context.Context, now time.Time) error {
	executor, err := postgres.Executor(ctx, r.pool)
	if err != nil {
		return err
	}
	_, err = queries.New(executor).Retain(ctx, postgres.Timestamptz(now))
	return err
}

func emailFromRow(row queries.FindRow) (Email, error) {
	key, err := tenant.Parse(row.Tenant)
	if err != nil {
		return Email{}, err
	}
	result := Email{
		ID:        row.ID.Bytes,
		Tenant:    key,
		Template:  mail.Template(row.Template),
		Recipient: row.Recipient,
		Subject:   row.Subject,
		Body:      row.Body,
		CreatedAt: row.CreatedAt.Time,
	}
	if err := json.Unmarshal(row.Params, &result.Params); err != nil {
		return Email{}, err
	}
	switch Status(row.Status) {
	case StatusPending:
		result.Delivery = Pending{}
	case StatusHeld:
		result.Delivery = Held{}
	case StatusSent:
		result.Delivery = Sent{At: row.SentAt.Time, ProviderMessageID: row.ProviderMessageID.String}
	case StatusFailed:
		result.Delivery = Failed{Error: row.Error.String}
	default:
		return Email{}, fmt.Errorf("invalid email status %q", row.Status)
	}
	return result, nil
}

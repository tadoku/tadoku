package email

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
)

type Service struct{}

var (
	_ Writer = (*Service)(nil)
	_ Reader = (*Service)(nil)
)

func NewService() *Service { return &Service{} }
func (*Service) Write(context.Context, mail.Message) (WriteResult, error) {
	return WriteResult{}, ErrNotImplemented
}
func (*Service) List(context.Context, Filter) (Page, error)     { return Page{}, ErrNotImplemented }
func (*Service) Find(context.Context, uuid.UUID) (Email, error) { return Email{}, ErrNotImplemented }
func (*Service) Send(context.Context, uuid.UUID) error          { return ErrNotImplemented }
func (*Service) Retain(context.Context) error                   { return ErrNotImplemented }

package logquery

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type LogRepository interface {
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
}

type Service interface {
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
}

type service struct {
	r        LogRepository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(r LogRepository, clock domain.Clock) Service {
	return &service{
		r:        r,
		validate: validator.New(),
		clock:    clock,
	}
}

type Language struct {
	Code string
	Name string
}

type Activity struct {
	ID      int32
	Name    string
	Default bool
}

type Unit struct {
	ID            uuid.UUID
	LogActivityID int
	Name          string
	Modifier      float32
	LanguageCode  *string
}

type Tag struct {
	ID            uuid.UUID
	LogActivityID int
	Name          string
}

type FetchLogConfigurationOptionsResponse struct {
	Languages  []Language
	Activities []Activity
	Units      []Unit
	Tags       []Tag
}

func (s *service) FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	return s.r.FetchLogConfigurationOptions(ctx)
}

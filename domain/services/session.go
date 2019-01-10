package services

import (
	"github.com/tadoku/api/domain"
)

type SessionService interface {
	Login(ctx domain.Context) error
	Register(ctx domain.Context) error
}

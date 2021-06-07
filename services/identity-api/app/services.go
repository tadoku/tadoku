package app

import (
	"github.com/tadoku/tadoku/services/identity-api/interfaces/services"
)

// Services is a collection of all services
type Services struct {
	Health  services.HealthService
	Session services.SessionService
	User    services.UserService
}

// NewServices initializes all interactors
func NewServices(i *Interactors, sessionCookieName string) *Services {
	return &Services{
		Health:  services.NewHealthService(),
		Session: services.NewSessionService(i.Session, sessionCookieName),
		User:    services.NewUserService(i.User),
	}
}

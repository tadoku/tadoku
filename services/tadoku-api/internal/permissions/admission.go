package permissions

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Admission interface {
	admission()
}

type Admitted struct{}
type Banned struct{}

type AdmittedBanUnknown struct {
	Err error
}

type Unavailable struct {
	Err error
}

func (Admitted) admission()           {}
func (Banned) admission()             {}
func (AdmittedBanUnknown) admission() {}
func (Unavailable) admission()        {}

func (c *Checker) Admit(ctx context.Context) Admission {
	if _, err := rolesObject(ctx); err != nil {
		return Unavailable{Err: err}
	}

	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return Admitted{}
	}

	banned, err := c.CheckBanned(ctx, user.Subject)
	if err != nil {
		return AdmittedBanUnknown{Err: err}
	}
	if banned {
		return Banned{}
	}
	return Admitted{}
}

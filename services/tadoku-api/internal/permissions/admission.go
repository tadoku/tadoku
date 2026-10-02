package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Admission interface {
	admission()
}

type Admitted struct{}
type Banned struct{}
type NoAccess struct{}

type AdmittedBanUnknown struct {
	Err error
}

type Unavailable struct {
	Err error
}

func (Admitted) admission()           {}
func (Banned) admission()             {}
func (NoAccess) admission()           {}
func (AdmittedBanUnknown) admission() {}
func (Unavailable) admission()        {}

func (c *Checker) Admit(ctx context.Context) Admission {
	object, err := rolesObject(ctx)
	if err != nil {
		return Unavailable{Err: err}
	}
	key, _ := tenant.FromContext(ctx)
	production := key == tenant.Production()
	user := identity.FromContext(ctx)

	if user == nil || user.Subject == "" || user.Subject == "guest" {
		if production {
			return Admitted{}
		}
		return NoAccess{}
	}

	if production {
		banned, err := c.CheckBanned(ctx, user.Subject)
		if err != nil {
			return AdmittedBanUnknown{Err: err}
		}
		if banned {
			return Banned{}
		}
		return Admitted{}
	}

	if c == nil || c.client == nil {
		return Unavailable{Err: errx.NewUnavailableError("permissions unavailable", nil)}
	}
	checks := []ketoclient.PermissionCheck{
		{
			Namespace: "app",
			Object:    object,
			Relation:  "is_banned",
			Subject:   ketoclient.Subject{ID: user.Subject},
		},
		{
			Namespace: "app",
			Object:    object,
			Relation:  "access",
			Subject:   ketoclient.Subject{ID: user.Subject},
		},
	}
	results := c.client.CheckPermissions(ctx, checks)
	if err := ctx.Err(); err != nil {
		return Unavailable{Err: errx.NewUnavailableError("check tenant access", err)}
	}
	for _, result := range results {
		if result.Err != nil {
			return Unavailable{Err: errx.NewUnavailableError("check tenant access", result.Err)}
		}
	}

	if results[0].Allowed {
		return Banned{}
	}
	if !results[1].Allowed {
		return NoAccess{}
	}
	return Admitted{}
}

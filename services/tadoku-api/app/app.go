// Package app authorizes requests and composes concrete feature services.
package app

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/tadoku/tadoku/services/tadoku-api/features/access"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

var ErrForbidden = errors.New("forbidden")

// Principal contains verified credential facts, never caller-supplied headers.
type Principal struct {
	Subject  string
	Service  bool
	Audience []string
}

type Application struct {
	content *content.Service
	access  *access.Service
	logger  *slog.Logger
}

func New(content *content.Service, access *access.Service, logger *slog.Logger) *Application {
	return &Application{content: content, access: access, logger: logger}
}

func (a *Application) ActiveAnnouncements(ctx context.Context, principal Principal, namespace string) ([]content.Announcement, error) {
	if principal.Service {
		// Consolidating processes does not consolidate credential audiences.
		if !slices.Contains(principal.Audience, "content-api") {
			return nil, ErrForbidden
		}
	} else {
		banned, err := a.access.Banned(ctx, principal.Subject)
		if err != nil {
			// Preserve Content's existing fail-open policy for this public read.
			// Do not reuse this rule for future administrative operations.
			a.logger.WarnContext(ctx, "authorization unavailable for public content read", "error", err)
		} else if banned {
			return nil, ErrForbidden
		}
	}
	return a.content.ActiveAnnouncements(ctx, namespace)
}

// Package app composes concrete feature services into application operations.
package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Application struct {
	announcements *announcements.Service
	audit         *audit.Service
	authorization *authz.Service
	contests      *contests.Service
	leaderboard   *leaderboard.Service
	languages     *languages.Service
	logs          *logs.Service
	pages         *pages.Service
	posts         *posts.Service
	profile       *profile.Service
	scoring       *scoring.Service
	featureFlags  *featureflags.Service
	db            *pgxpool.Pool
	permissions   *permissions.Checker
}

func New(announcements *announcements.Service, audit *audit.Service, authorization *authz.Service, contests *contests.Service, leaderboard *leaderboard.Service, languages *languages.Service, logs *logs.Service, pages *pages.Service, posts *posts.Service, profile *profile.Service, scoring *scoring.Service, featureFlags *featureflags.Service, db *pgxpool.Pool, permissions *permissions.Checker) *Application {
	return &Application{
		announcements: announcements,
		audit:         audit,
		authorization: authorization,
		contests:      contests,
		leaderboard:   leaderboard,
		languages:     languages,
		logs:          logs,
		pages:         pages,
		posts:         posts,
		profile:       profile,
		scoring:       scoring,
		featureFlags:  featureFlags,
		db:            db,
		permissions:   permissions,
	}
}

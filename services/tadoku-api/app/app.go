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

// Dependencies holds the concrete collaborators composed by the application.
type Dependencies struct {
	Announcements *announcements.Service
	Audit         *audit.Service
	Authorization *authz.Service
	Contests      *contests.Service
	Leaderboard   *leaderboard.Service
	Languages     *languages.Service
	Logs          *logs.Service
	Pages         *pages.Service
	Posts         *posts.Service
	Profile       *profile.Service
	Scoring       *scoring.Service
	FeatureFlags  *featureflags.Service
	DB            *pgxpool.Pool
	Permissions   *permissions.Checker
}

func New(deps Dependencies) *Application {
	return &Application{
		announcements: deps.Announcements,
		audit:         deps.Audit,
		authorization: deps.Authorization,
		contests:      deps.Contests,
		leaderboard:   deps.Leaderboard,
		languages:     deps.Languages,
		logs:          deps.Logs,
		pages:         deps.Pages,
		posts:         deps.Posts,
		profile:       deps.Profile,
		scoring:       deps.Scoring,
		featureFlags:  deps.FeatureFlags,
		db:            deps.DB,
		permissions:   deps.Permissions,
	}
}

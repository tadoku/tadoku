package api

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz/tokens"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Application struct {
	tokens        *tokens.Service
	deployment    tenant.Deployment
	email         *email.Service
	jobqueue      *jobqueue.Service
	announcements *announcements.Service
	audit         *audit.Service
	authorization *authz.Service
	contests      *contests.Service
	contestEvents *contestevents.Service
	leaderboard   *leaderboard.Service
	languages     *languages.Service
	logs          *logs.Service
	pages         *pages.Service
	posts         *posts.Service
	profile       *profile.Service
	scoring       *scoring.Service
	featureFlags  *featureflags.Evaluator
	db            *pgxpool.Pool
	identity      *identity.Service
	permissions   *permissions.Checker
	roles         *authz.Roles
}

type Dependencies struct {
	Tokens        *tokens.Service
	Deployment    tenant.Deployment
	Email         *email.Service
	JobQueue      *jobqueue.Service
	Announcements *announcements.Service
	Audit         *audit.Service
	Authorization *authz.Service
	Contests      *contests.Service
	ContestEvents *contestevents.Service
	Leaderboard   *leaderboard.Service
	Languages     *languages.Service
	Logs          *logs.Service
	Pages         *pages.Service
	Posts         *posts.Service
	Profile       *profile.Service
	Scoring       *scoring.Service
	FeatureFlags  *featureflags.Evaluator
	DB            *pgxpool.Pool
	Identity      *identity.Service
	Permissions   *permissions.Checker
	Roles         *authz.Roles
}

func New(deps Dependencies) *Application {
	return &Application{
		tokens:        deps.Tokens,
		deployment:    deps.Deployment,
		email:         deps.Email,
		jobqueue:      deps.JobQueue,
		announcements: deps.Announcements,
		audit:         deps.Audit,
		authorization: deps.Authorization,
		contests:      deps.Contests,
		contestEvents: deps.ContestEvents,
		leaderboard:   deps.Leaderboard,
		languages:     deps.Languages,
		logs:          deps.Logs,
		pages:         deps.Pages,
		posts:         deps.Posts,
		profile:       deps.Profile,
		scoring:       deps.Scoring,
		featureFlags:  deps.FeatureFlags,
		db:            deps.DB,
		identity:      deps.Identity,
		permissions:   deps.Permissions,
		roles:         deps.Roles,
	}
}

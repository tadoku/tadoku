package authz

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const appRoot = "root"

var Namespace = permissions.Define(
	"app",
	permissions.Relations{
		"admins": {permissions.Users},
		"banned": {permissions.Users},
	},
	permissions.Permits{
		"admin":     permissions.Related("admins"),
		"is_banned": permissions.Related("banned"),
	},
)

var (
	adminCheck  = Namespace.Check(appRoot, "admin")
	bannedCheck = Namespace.Check(appRoot, "is_banned")
)

func adminRelationship(userID uuid.UUID) permissions.Relationship {
	return Namespace.UserRelationship(appRoot, "admins", userID)
}

func banRelationship(userID uuid.UUID) permissions.Relationship {
	return Namespace.UserRelationship(appRoot, "banned", userID)
}

// Roles answers and changes the administrator and ban roles stored on the context tenant's app:root object, which
// Namespace defines. Build the checker and relationships from a model that includes Namespace.
type Roles struct {
	checker       *permissions.Checker
	relationships *permissions.Relationships
}

func NewRoles(checker *permissions.Checker, relationships *permissions.Relationships) *Roles {
	return &Roles{checker: checker, relationships: relationships}
}

// RequireAdmin returns unauthorized when nobody is signed in, forbidden for a user who is not an administrator, and
// unavailable when the check fails or the request's ban state is unknown.
func (r *Roles) RequireAdmin(ctx context.Context) error {
	return r.checker.Require(ctx, adminCheck)
}

func (r *Roles) IsAdmin(ctx context.Context) (bool, error) {
	user := requestidentity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return false, nil
	}
	return r.checker.Allowed(ctx, adminCheck)
}

func (r *Roles) IsAdminOrFalse(ctx context.Context) bool {
	allowed, err := r.IsAdmin(ctx)
	return err == nil && allowed
}

// RequireProductionAdmin guards writes to data shared by every tenant, such as languages and log units.
func (r *Roles) RequireProductionAdmin(ctx context.Context) error {
	if err := r.RequireAdmin(ctx); err != nil {
		return err
	}

	key, _ := tenant.FromContext(ctx)
	if key != tenant.Production() {
		return errx.NewForbiddenError("shared data is writable only from the production tenant")
	}
	return nil
}

// IsProductionAdmin evaluates tadoku/prod whatever tenant the context carries. A subject that is not a UUID is not an
// administrator.
func (r *Roles) IsProductionAdmin(ctx context.Context, subjectID string) (bool, error) {
	userID, err := uuid.Parse(subjectID)
	if err != nil {
		return false, nil
	}
	return r.checker.AllowedFor(tenant.WithKey(ctx, tenant.Production()), adminCheck, userID)
}

// Admit checks the signed-in user's ban with one query. A context without a tenant is Unavailable, even for guests.
func (r *Roles) Admit(ctx context.Context) permissions.Admission {
	if _, ok := tenant.FromContext(ctx); !ok {
		return permissions.Unavailable{Err: errx.NewUnavailableError("permissions require a tenant", nil)}
	}

	userID, ok := requestidentity.ActorID(ctx)
	if !ok {
		return permissions.Admitted{}
	}
	banned, err := r.checker.AllowedFor(ctx, bannedCheck, userID)
	if err != nil {
		return permissions.AdmittedBanUnknown{Err: err}
	}
	if banned {
		return permissions.Banned{}
	}
	return permissions.Admitted{}
}

// UserRoles reads each user's stored role in the context tenant with one query. RoleAdmin wins over RoleBanned, and a
// user with neither is RoleUser.
func (r *Roles) UserRoles(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]Role, error) {
	candidates := make([]permissions.Relationship, 0, 2*len(userIDs))
	for _, userID := range userIDs {
		candidates = append(candidates, adminRelationship(userID), banRelationship(userID))
	}
	stored, err := r.relationships.Stored(ctx, candidates)
	if err != nil {
		return nil, err
	}

	roles := make(map[uuid.UUID]Role, len(userIDs))
	for _, userID := range userIDs {
		switch {
		case stored[adminRelationship(userID)]:
			roles[userID] = RoleAdmin
		case stored[banRelationship(userID)]:
			roles[userID] = RoleBanned
		default:
			roles[userID] = RoleUser
		}
	}
	return roles, nil
}

// SetBanned stores or removes the user's ban in the context tenant and joins the context's transaction.
func (r *Roles) SetBanned(ctx context.Context, userID uuid.UUID, banned bool) error {
	if banned {
		return r.relationships.Grant(ctx, banRelationship(userID))
	}
	return r.relationships.Revoke(ctx, banRelationship(userID))
}

// GrantAdmin makes the user an administrator of the context tenant and joins the context's transaction. Build
// relationships from a model that includes Namespace.
func GrantAdmin(ctx context.Context, relationships *permissions.Relationships, userID uuid.UUID) error {
	return relationships.Grant(ctx, adminRelationship(userID))
}

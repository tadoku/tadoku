package permissions

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

const (
	Group   ObjectType = "group"
	Contest ObjectType = "contest"
	Log     ObjectType = "log"
)

var testModel = Model{
	App: model[App],
	Group: {
		Relations: Relations{
			"members": {Users, SetOf(Group, "members")},
		},
	},
	Contest: {
		Relations: Relations{
			"owners":     {Users},
			"moderators": {Users, SetOf(Group, "members")},
		},
		Permits: Permits{
			"edit":     AnyOf(Related("owners"), AppPermit("admin")),
			"moderate": AnyOf(Allowed("edit"), Related("moderators")),
		},
	},
	Log: {
		Relations: Relations{
			"contests": {ObjectOf(Contest)},
		},
		Permits: Permits{
			"hide": AnyOf(AppPermit("admin"), Via("contests", "moderate")),
		},
	},
}

func editContest(contestID uuid.UUID) Check {
	return Check{object: contestObject(contestID), permit: "edit"}
}

func moderateContest(contestID uuid.UUID) Check {
	return Check{object: contestObject(contestID), permit: "moderate"}
}

func hideLog(logID uuid.UUID) Check {
	return Check{object: logObject(logID), permit: "hide"}
}

func contestOwner(contestID, userID uuid.UUID) Relationship {
	return Relationship{object: contestObject(contestID), relation: "owners", subject: userSubject(userID)}
}

func contestModerator(contestID, userID uuid.UUID) Relationship {
	return Relationship{object: contestObject(contestID), relation: "moderators", subject: userSubject(userID)}
}

func contestModeratorGroup(contestID, groupID uuid.UUID) Relationship {
	return Relationship{object: contestObject(contestID), relation: "moderators", subject: groupMembers(groupID)}
}

func groupMember(groupID, userID uuid.UUID) Relationship {
	return Relationship{object: groupObject(groupID), relation: "members", subject: userSubject(userID)}
}

func subgroup(groupID, memberGroupID uuid.UUID) Relationship {
	return Relationship{object: groupObject(groupID), relation: "members", subject: groupMembers(memberGroupID)}
}

func logCountsFor(logID, contestID uuid.UUID) Relationship {
	return Relationship{
		object:   logObject(logID),
		relation: "contests",
		subject:  subjectSet{object: contestObject(contestID)},
	}
}

func contestObject(id uuid.UUID) object { return object{objectType: Contest, id: id.String()} }
func groupObject(id uuid.UUID) object   { return object{objectType: Group, id: id.String()} }
func logObject(id uuid.UUID) object     { return object{objectType: Log, id: id.String()} }

func groupMembers(groupID uuid.UUID) subjectSet {
	return subjectSet{object: groupObject(groupID), relation: "members"}
}

func newTestDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

func productionContext(t *testing.T) context.Context {
	t.Helper()

	return tenant.WithKey(t.Context(), tenant.Production())
}

func registerTestTenant(t *testing.T, db *testpostgres.Database, raw string) context.Context {
	t.Helper()

	key, err := tenant.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into tenants (key, kind) values ($1, 'test')`, key.String())
	if err != nil {
		t.Fatal(err)
	}
	return tenant.WithKey(t.Context(), key)
}

func grantAll(t *testing.T, ctx context.Context, grants *Grants, relationships ...Relationship) {
	t.Helper()

	for _, relationship := range relationships {
		if err := grants.Grant(ctx, relationship); err != nil {
			t.Fatalf("grant %+v: %v", relationship, err)
		}
	}
}

// storedRelationships reads one tenant's rows through the owner pool as object#relation@subject#subject_relation.
func storedRelationships(t *testing.T, db *testpostgres.Database, key string) []string {
	t.Helper()

	rows, err := db.Pool.Query(t.Context(), `select object_type || ':' || object_id || '#' || relation || '@'
		|| subject_type || ':' || subject_id || '#' || subject_relation
		from permission_relationships
		where tenant = $1
		order by 1`,
		key,
	)
	if err != nil {
		t.Fatal(err)
	}

	stored, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

type statementLog struct {
	mu         sync.Mutex
	statements []string
}

func (l *statementLog) record(sql string) {
	if !strings.Contains(sql, "permission_relationships") {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.statements = append(l.statements, sql)
}

func (l *statementLog) take() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	statements := l.statements
	l.statements = nil
	return statements
}

func (l *statementLog) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l.record(data.SQL)
	return ctx
}

func (l *statementLog) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (l *statementLog) TraceBatchStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceBatchStartData) context.Context {
	return ctx
}

func (l *statementLog) TraceBatchQuery(_ context.Context, _ *pgx.Conn, data pgx.TraceBatchQueryData) {
	l.record(data.SQL)
}

func (l *statementLog) TraceBatchEnd(context.Context, *pgx.Conn, pgx.TraceBatchEndData) {}

// newLoggedAppPool opens an application-role pool that records every permission_relationships statement.
func newLoggedAppPool(t *testing.T, db *testpostgres.Database) (*pgxpool.Pool, *statementLog) {
	t.Helper()

	statements := &statementLog{}
	config := db.AppPool.Config()
	config.ConnConfig.Tracer = statements

	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool, statements
}

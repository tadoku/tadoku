package authz

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

const (
	siteType    permissions.ObjectType = "site"
	groupType   permissions.ObjectType = "group"
	contestType permissions.ObjectType = "contest"
	logType     permissions.ObjectType = "log"
)

const siteRoot = "root"

var (
	sites = permissions.Define(
		siteType,
		permissions.Relations{
			"admins": {permissions.Users},
			"banned": {permissions.Users},
		},
		permissions.Permits{"admin": permissions.Related("admins")},
	)
	groups = permissions.Define(
		groupType,
		permissions.Relations{"members": {permissions.Users, permissions.SetOf(groupType, "members")}},
		nil,
	)
	contests = permissions.Define(
		contestType,
		permissions.Relations{
			"owners":     {permissions.Users},
			"moderators": {permissions.Users, permissions.SetOf(groupType, "members")},
		},
		permissions.Permits{
			"edit": permissions.AnyOf(
				permissions.Related("owners"),
				permissions.PermitOn(siteType, siteRoot, "admin"),
			),
			"moderate": permissions.AnyOf(permissions.Allowed("edit"), permissions.Related("moderators")),
		},
	)
	logs = permissions.Define(
		logType,
		permissions.Relations{"contests": {permissions.ObjectOf(contestType)}},
		permissions.Permits{"hide": permissions.AnyOf(
			permissions.PermitOn(siteType, siteRoot, "admin"),
			permissions.Via("contests", "moderate"),
		)},
	)
	testModel = mustModel(sites, groups, contests, logs)
	siteAdmin = sites.Check(siteRoot, "admin")
)

func mustModel(namespaces ...permissions.Namespace) permissions.Model {
	model, err := permissions.NewModel(namespaces...)
	if err != nil {
		panic(err)
	}
	return model
}

func siteAdminRelationship(userID uuid.UUID) permissions.Relationship {
	return sites.UserRelationship(siteRoot, "admins", userID)
}

func siteBanRelationship(userID uuid.UUID) permissions.Relationship {
	return sites.UserRelationship(siteRoot, "banned", userID)
}

func editContest(contestID uuid.UUID) permissions.Check {
	return contests.Check(contestID.String(), "edit")
}

func moderateContest(contestID uuid.UUID) permissions.Check {
	return contests.Check(contestID.String(), "moderate")
}

func hideLog(logID uuid.UUID) permissions.Check {
	return logs.Check(logID.String(), "hide")
}

func contestOwner(contestID, userID uuid.UUID) permissions.Relationship {
	return contests.UserRelationship(contestID.String(), "owners", userID)
}

func contestModerator(contestID, userID uuid.UUID) permissions.Relationship {
	return contests.UserRelationship(contestID.String(), "moderators", userID)
}

func groupMember(groupID, userID uuid.UUID) permissions.Relationship {
	return groups.UserRelationship(groupID.String(), "members", userID)
}

func contestModeratorGroup(contestID, groupID uuid.UUID) permissions.Tuple {
	return groupMembersTuple(contestType, contestID, "moderators", groupID)
}

func subgroup(groupID, memberGroupID uuid.UUID) permissions.Tuple {
	return groupMembersTuple(groupType, groupID, "members", memberGroupID)
}

func groupMembersTuple(
	objectType permissions.ObjectType,
	objectID uuid.UUID,
	relation string,
	groupID uuid.UUID,
) permissions.Tuple {
	return permissions.Tuple{
		ObjectType:      string(objectType),
		ObjectID:        objectID.String(),
		Relation:        relation,
		SubjectType:     string(groupType),
		SubjectID:       groupID.String(),
		SubjectRelation: "members",
	}
}

func logCountsFor(logID, contestID uuid.UUID) permissions.Tuple {
	return permissions.Tuple{
		ObjectType:  string(logType),
		ObjectID:    logID.String(),
		Relation:    "contests",
		SubjectType: string(contestType),
		SubjectID:   contestID.String(),
	}
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

func closedPool(t *testing.T, db *testpostgres.Database) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.NewWithConfig(t.Context(), db.AppPool.Config())
	if err != nil {
		t.Fatal(err)
	}
	pool.Close()
	return pool
}

func grantAll(
	t *testing.T,
	ctx context.Context,
	writer *permissions.Relationships,
	relationships ...permissions.Relationship,
) {
	t.Helper()

	for _, relationship := range relationships {
		if err := writer.Grant(ctx, relationship); err != nil {
			t.Fatalf("grant %+v: %v", relationship, err)
		}
	}
}

// insertAll stores subject sets and linked objects, which no namespace builds yet.
func insertAll(t *testing.T, ctx context.Context, repository *Repository, tuples ...permissions.Tuple) {
	t.Helper()

	for _, tuple := range tuples {
		if err := repository.Insert(ctx, tuple); err != nil {
			t.Fatalf("insert %+v: %v", tuple, err)
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
	// onBatchEnd runs once a batch's results are read, before the query call returns.
	onBatchEnd func()
}

func (l *statementLog) record(sql string) {
	if !strings.Contains(sql, "permission_relationships") {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.statements = append(l.statements, sql)
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

func (l *statementLog) TraceBatchEnd(context.Context, *pgx.Conn, pgx.TraceBatchEndData) {
	if l.onBatchEnd != nil {
		l.onBatchEnd()
	}
}

func (l *statementLog) take() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	statements := l.statements
	l.statements = nil
	return statements
}

// newLoggedPool opens an application-role pool that records every permission_relationships query.
func newLoggedPool(t *testing.T, db *testpostgres.Database) (*pgxpool.Pool, *statementLog) {
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

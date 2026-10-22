package permissions

import (
	"context"
	"errors"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlcpermissions "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const userSubjectType = "user"

// store queries join the context's transaction, so a check sees that transaction's own grants.
type store struct {
	db *pgxpool.Pool
}

func (s store) queries(ctx context.Context) (*sqlcpermissions.Queries, error) {
	db, err := postgres.Executor(ctx, s.db)
	if errors.Is(err, postgres.ErrNoTenant) {
		return nil, errx.NewUnavailableError("permissions require a tenant", err)
	}
	if err != nil {
		return nil, err
	}
	return sqlcpermissions.New(db), nil
}

// subjects reports whether user is stored in the relation and returns the subject sets and linked objects stored in it.
func (s store) subjects(ctx context.Context, obj object, relation string, user uuid.UUID) (bool, []subjectSet, error) {
	queries, err := s.queries(ctx)
	if err != nil {
		return false, nil, err
	}

	rows, err := queries.RelationshipSubjects(ctx, sqlcpermissions.RelationshipSubjectsParams{
		ObjectType: string(obj.objectType),
		ObjectID:   obj.id,
		Relation:   relation,
		UserID:     user.String(),
	})
	if err != nil {
		return false, nil, err
	}

	sets := make([]subjectSet, 0, len(rows))
	for _, row := range rows {
		if row.SubjectType == userSubjectType {
			return true, nil, nil
		}
		sets = append(sets, subjectSet{
			object:   object{objectType: ObjectType(row.SubjectType), id: row.SubjectID},
			relation: row.SubjectRelation,
		})
	}
	return false, sets, nil
}

// appRelationships returns the stored relationships between the users and app:root.
func (s store) appRelationships(ctx context.Context, users []uuid.UUID) (map[Relationship]bool, error) {
	queries, err := s.queries(ctx)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}

	ids := make([]string, len(users))
	for i, user := range users {
		ids[i] = user.String()
	}
	rows, err := queries.UserRelations(ctx, sqlcpermissions.UserRelationsParams{
		Relations: slices.Collect(maps.Keys(model[App].Relations)),
		UserIds:   ids,
	})
	if err != nil {
		return nil, err
	}

	stored := make(map[Relationship]bool, len(rows))
	for _, row := range rows {
		user, err := uuid.Parse(row.SubjectID)
		if err != nil {
			return nil, err
		}
		stored[Relationship{object: appRoot, relation: row.Relation, subject: userSubject(user)}] = true
	}
	return stored, nil
}

// insert and delete accept only relationships that passed Model.validateRelationship.
func (s store) insert(ctx context.Context, relationship Relationship) error {
	queries, err := s.queries(ctx)
	if err != nil {
		return err
	}

	_, err = queries.InsertRelationship(ctx, relationshipRow(relationship))
	return err
}

func (s store) delete(ctx context.Context, relationship Relationship) error {
	queries, err := s.queries(ctx)
	if err != nil {
		return err
	}

	_, err = queries.DeleteRelationship(ctx, sqlcpermissions.DeleteRelationshipParams(relationshipRow(relationship)))
	return err
}

func relationshipRow(relationship Relationship) sqlcpermissions.InsertRelationshipParams {
	row := sqlcpermissions.InsertRelationshipParams{
		ObjectType: string(relationship.object.objectType),
		ObjectID:   relationship.object.id,
		Relation:   relationship.relation,
	}

	switch subject := relationship.subject.(type) {
	case userSubject:
		row.SubjectType = userSubjectType
		row.SubjectID = uuid.UUID(subject).String()
	case subjectSet:
		row.SubjectType = string(subject.object.objectType)
		row.SubjectID = subject.object.id
		row.SubjectRelation = subject.relation
	}
	return row
}

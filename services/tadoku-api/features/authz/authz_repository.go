package authz

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// Repository stores every feature's permission relationship tuples in permission_relationships, in the context
// tenant. Each call joins the context's transaction when that transaction was begun on the same pool.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) queries(ctx context.Context) (*queries.Queries, error) {
	db, err := postgres.Executor(ctx, r.db)
	if errors.Is(err, postgres.ErrNoTenant) {
		return nil, errx.NewUnavailableError("permissions require a tenant", err)
	}
	if err != nil {
		return nil, err
	}
	return queries.New(db), nil
}

func (r *Repository) Subjects(
	ctx context.Context,
	objectType, objectID, relation, userID string,
) ([]permissions.Tuple, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.RelationshipSubjects(ctx, queries.RelationshipSubjectsParams{
		ObjectType: objectType,
		ObjectID:   objectID,
		Relation:   relation,
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	tuples := make([]permissions.Tuple, len(rows))
	for i, row := range rows {
		tuples[i] = permissions.Tuple{
			ObjectType:      objectType,
			ObjectID:        objectID,
			Relation:        relation,
			SubjectType:     row.SubjectType,
			SubjectID:       row.SubjectID,
			SubjectRelation: row.SubjectRelation,
		}
	}
	return tuples, nil
}

func (r *Repository) Insert(ctx context.Context, tuple permissions.Tuple) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.InsertRelationship(ctx, queries.InsertRelationshipParams(tuple))
	return err
}

func (r *Repository) Delete(ctx context.Context, tuple permissions.Tuple) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.DeleteRelationship(ctx, queries.DeleteRelationshipParams(tuple))
	return err
}

func (r *Repository) Stored(ctx context.Context, tuples []permissions.Tuple) ([]permissions.Tuple, error) {
	q, err := r.queries(ctx)
	if err != nil {
		return nil, err
	}
	if len(tuples) == 0 {
		return nil, nil
	}

	var wanted queries.StoredRelationshipsParams
	for _, tuple := range tuples {
		wanted.ObjectTypes = append(wanted.ObjectTypes, tuple.ObjectType)
		wanted.ObjectIds = append(wanted.ObjectIds, tuple.ObjectID)
		wanted.Relations = append(wanted.Relations, tuple.Relation)
		wanted.SubjectTypes = append(wanted.SubjectTypes, tuple.SubjectType)
		wanted.SubjectIds = append(wanted.SubjectIds, tuple.SubjectID)
		wanted.SubjectRelations = append(wanted.SubjectRelations, tuple.SubjectRelation)
	}
	rows, err := q.StoredRelationships(ctx, wanted)
	if err != nil {
		return nil, err
	}

	stored := make([]permissions.Tuple, len(rows))
	for i, row := range rows {
		stored[i] = permissions.Tuple(row)
	}
	return stored, nil
}

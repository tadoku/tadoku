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

func (r *Repository) Insert(ctx context.Context, tuples []permissions.Tuple) error {
	q, err := r.queries(ctx)
	if err != nil || len(tuples) == 0 {
		return err
	}

	_, err = q.InsertRelationships(ctx, queries.InsertRelationshipsParams(tupleColumns(tuples)))
	return err
}

func (r *Repository) Delete(ctx context.Context, tuples []permissions.Tuple) error {
	q, err := r.queries(ctx)
	if err != nil || len(tuples) == 0 {
		return err
	}

	_, err = q.DeleteRelationships(ctx, queries.DeleteRelationshipsParams(tupleColumns(tuples)))
	return err
}

func (r *Repository) DeleteObject(ctx context.Context, objectType, objectID string) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.DeleteObjectRelationships(ctx, queries.DeleteObjectRelationshipsParams{
		ObjectType: objectType,
		ObjectID:   objectID,
	})
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, userID string) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.DeleteUserRelationships(ctx, userID)
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

	rows, err := q.StoredRelationships(ctx, tupleColumns(tuples))
	if err != nil {
		return nil, err
	}

	stored := make([]permissions.Tuple, len(rows))
	for i, row := range rows {
		stored[i] = permissions.Tuple(row)
	}
	return stored, nil
}

func tupleColumns(tuples []permissions.Tuple) queries.StoredRelationshipsParams {
	var columns queries.StoredRelationshipsParams
	for _, tuple := range tuples {
		columns.ObjectTypes = append(columns.ObjectTypes, tuple.ObjectType)
		columns.ObjectIds = append(columns.ObjectIds, tuple.ObjectID)
		columns.Relations = append(columns.Relations, tuple.Relation)
		columns.SubjectTypes = append(columns.SubjectTypes, tuple.SubjectType)
		columns.SubjectIds = append(columns.SubjectIds, tuple.SubjectID)
		columns.SubjectRelations = append(columns.SubjectRelations, tuple.SubjectRelation)
	}
	return columns
}

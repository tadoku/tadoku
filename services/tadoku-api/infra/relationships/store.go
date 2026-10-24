package relationships

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/relationships"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const UserType = "user"

type Object struct {
	Type string
	ID   string
}

// Subject is a user (Type UserType, ID the user's UUID), a subject set (Relation names the set) or a linked object
// (empty Relation).
type Subject struct {
	Type     string
	ID       string
	Relation string
}

type Tuple struct {
	Object   Object
	Relation string
	Subject  Subject
}

// Store reads and writes permission_relationships in the context tenant. Each call joins the context's
// transaction when that transaction was begun on the same pool.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) queries(ctx context.Context) (*queries.Queries, error) {
	db, err := postgres.Executor(ctx, s.db)
	if errors.Is(err, postgres.ErrNoTenant) {
		return nil, errx.NewUnavailableError("permissions require a tenant", err)
	}
	if err != nil {
		return nil, err
	}
	return queries.New(db), nil
}

// Subjects returns userID's row when it is stored, and every stored subject in object#relation that is not a user.
func (s *Store) Subjects(ctx context.Context, object Object, relation string, userID string) ([]Subject, error) {
	q, err := s.queries(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := q.RelationshipSubjects(ctx, queries.RelationshipSubjectsParams{
		ObjectType: object.Type,
		ObjectID:   object.ID,
		Relation:   relation,
		UserID:     userID,
	})
	if err != nil {
		return nil, err
	}

	subjects := make([]Subject, len(rows))
	for i, row := range rows {
		subjects[i] = Subject{Type: row.SubjectType, ID: row.SubjectID, Relation: row.SubjectRelation}
	}
	return subjects, nil
}

// Insert keeps an already stored tuple.
func (s *Store) Insert(ctx context.Context, tuple Tuple) error {
	q, err := s.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.InsertRelationship(ctx, insertParams(tuple))
	return err
}

// Delete succeeds when the tuple is not stored.
func (s *Store) Delete(ctx context.Context, tuple Tuple) error {
	q, err := s.queries(ctx)
	if err != nil {
		return err
	}

	_, err = q.DeleteRelationship(ctx, queries.DeleteRelationshipParams(insertParams(tuple)))
	return err
}

// Stored returns the given tuples that are stored, with one query. Without a tenant it fails even for no tuples.
func (s *Store) Stored(ctx context.Context, tuples []Tuple) ([]Tuple, error) {
	q, err := s.queries(ctx)
	if err != nil {
		return nil, err
	}
	if len(tuples) == 0 {
		return nil, nil
	}

	var wanted queries.StoredRelationshipsParams
	for _, tuple := range tuples {
		wanted.ObjectTypes = append(wanted.ObjectTypes, tuple.Object.Type)
		wanted.ObjectIds = append(wanted.ObjectIds, tuple.Object.ID)
		wanted.Relations = append(wanted.Relations, tuple.Relation)
		wanted.SubjectTypes = append(wanted.SubjectTypes, tuple.Subject.Type)
		wanted.SubjectIds = append(wanted.SubjectIds, tuple.Subject.ID)
		wanted.SubjectRelations = append(wanted.SubjectRelations, tuple.Subject.Relation)
	}
	rows, err := q.StoredRelationships(ctx, wanted)
	if err != nil {
		return nil, err
	}

	stored := make([]Tuple, len(rows))
	for i, row := range rows {
		stored[i] = Tuple{
			Object:   Object{Type: row.ObjectType, ID: row.ObjectID},
			Relation: row.Relation,
			Subject:  Subject{Type: row.SubjectType, ID: row.SubjectID, Relation: row.SubjectRelation},
		}
	}
	return stored, nil
}

func insertParams(tuple Tuple) queries.InsertRelationshipParams {
	return queries.InsertRelationshipParams{
		ObjectType:      tuple.Object.Type,
		ObjectID:        tuple.Object.ID,
		Relation:        tuple.Relation,
		SubjectType:     tuple.Subject.Type,
		SubjectID:       tuple.Subject.ID,
		SubjectRelation: tuple.Subject.Relation,
	}
}

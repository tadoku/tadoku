package permissions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/relationships"
)

const userSubjectType = relationships.UserType

// store calls join the context's transaction, so a check sees that transaction's own grants.
type store struct {
	tuples *relationships.Store
}

func newStore(db *pgxpool.Pool) store {
	return store{tuples: relationships.NewStore(db)}
}

// subjects reports whether user is stored in the relation and returns the subject sets and linked objects stored in it.
func (s store) subjects(ctx context.Context, obj object, relation string, user uuid.UUID) (bool, []subjectSet, error) {
	stored, err := s.tuples.Subjects(ctx, tupleObject(obj), relation, user.String())
	if err != nil {
		return false, nil, err
	}

	sets := make([]subjectSet, 0, len(stored))
	for _, subject := range stored {
		if subject.Type == userSubjectType {
			return true, nil, nil
		}
		sets = append(sets, subjectSet{
			object:   object{objectType: ObjectType(subject.Type), id: subject.ID},
			relation: subject.Relation,
		})
	}
	return false, sets, nil
}

// insert, delete and stored accept only relationships that passed Model.validateRelationship.
func (s store) insert(ctx context.Context, relationship Relationship) error {
	return s.tuples.Insert(ctx, tuple(relationship))
}

func (s store) delete(ctx context.Context, relationship Relationship) error {
	return s.tuples.Delete(ctx, tuple(relationship))
}

func (s store) stored(ctx context.Context, candidates []Relationship) (map[Relationship]bool, error) {
	byTuple := make(map[relationships.Tuple]Relationship, len(candidates))
	tuples := make([]relationships.Tuple, 0, len(candidates))
	for _, candidate := range candidates {
		t := tuple(candidate)
		byTuple[t] = candidate
		tuples = append(tuples, t)
	}

	rows, err := s.tuples.Stored(ctx, tuples)
	if err != nil {
		return nil, err
	}

	stored := make(map[Relationship]bool, len(rows))
	for _, row := range rows {
		stored[byTuple[row]] = true
	}
	return stored, nil
}

func tupleObject(obj object) relationships.Object {
	return relationships.Object{Type: string(obj.objectType), ID: obj.id}
}

func tuple(relationship Relationship) relationships.Tuple {
	t := relationships.Tuple{Object: tupleObject(relationship.object), Relation: relationship.relation}
	switch subject := relationship.subject.(type) {
	case userSubject:
		t.Subject = relationships.Subject{Type: userSubjectType, ID: uuid.UUID(subject).String()}
	case subjectSet:
		t.Subject = relationships.Subject{
			Type:     string(subject.object.objectType),
			ID:       subject.object.id,
			Relation: subject.relation,
		}
	}
	return t
}

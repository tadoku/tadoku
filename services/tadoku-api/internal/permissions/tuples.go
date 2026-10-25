package permissions

import (
	"context"

	"github.com/google/uuid"
)

const UserSubjectType = "user"

// Tuple is one stored relationship. A user subject has SubjectType UserSubjectType, the user's UUID as SubjectID and
// an empty SubjectRelation. A linked object also has an empty SubjectRelation.
type Tuple struct {
	ObjectType      string
	ObjectID        string
	Relation        string
	SubjectType     string
	SubjectID       string
	SubjectRelation string
}

// Tuples stores relationship tuples for the context tenant. Every call joins the transaction the context carries, so
// a check sees that transaction's own writes, and fails as unavailable without a tenant.
type Tuples interface {
	// Subjects returns the tuples stored in objectType:objectID#relation whose subject is userID or is not a user.
	Subjects(ctx context.Context, objectType, objectID, relation, userID string) ([]Tuple, error)
	// Insert keeps a tuple that is already stored.
	Insert(ctx context.Context, tuple Tuple) error
	// Delete succeeds when the tuple is not stored.
	Delete(ctx context.Context, tuple Tuple) error
	// Stored returns the given tuples that are stored, with at most one query. It fails without a tenant even for no
	// tuples.
	Stored(ctx context.Context, tuples []Tuple) ([]Tuple, error)
}

// storedSubjects reports whether user is stored in the relation and returns the subject sets and linked objects
// stored in it.
func storedSubjects(
	ctx context.Context,
	tuples Tuples,
	obj Object,
	relation string,
	user uuid.UUID,
) (bool, []subjectSet, error) {
	stored, err := tuples.Subjects(ctx, string(obj.objectType), obj.id, relation, user.String())
	if err != nil {
		return false, nil, err
	}

	sets := make([]subjectSet, 0, len(stored))
	for _, tuple := range stored {
		if tuple.SubjectType == UserSubjectType {
			return true, nil, nil
		}
		sets = append(sets, subjectSet{
			object:   Object{objectType: ObjectType(tuple.SubjectType), id: tuple.SubjectID},
			relation: tuple.SubjectRelation,
		})
	}
	return false, sets, nil
}

func storedRelationships(ctx context.Context, tuples Tuples, candidates []Relationship) (map[Relationship]bool, error) {
	byTuple := make(map[Tuple]Relationship, len(candidates))
	wanted := make([]Tuple, 0, len(candidates))
	for _, candidate := range candidates {
		tuple := tupleOf(candidate)
		byTuple[tuple] = candidate
		wanted = append(wanted, tuple)
	}

	rows, err := tuples.Stored(ctx, wanted)
	if err != nil {
		return nil, err
	}

	stored := make(map[Relationship]bool, len(rows))
	for _, row := range rows {
		stored[byTuple[row]] = true
	}
	return stored, nil
}

// tupleOf accepts only relationships that passed Model.validateRelationship.
func tupleOf(relationship Relationship) Tuple {
	tuple := Tuple{
		ObjectType: string(relationship.object.objectType),
		ObjectID:   relationship.object.id,
		Relation:   relationship.relation,
	}
	switch subject := relationship.subject.(type) {
	case userSubject:
		tuple.SubjectType = UserSubjectType
		tuple.SubjectID = uuid.UUID(subject).String()
	case Object:
		tuple.SubjectType = string(subject.objectType)
		tuple.SubjectID = subject.id
	case subjectSet:
		tuple.SubjectType = string(subject.object.objectType)
		tuple.SubjectID = subject.object.id
		tuple.SubjectRelation = subject.relation
	}
	return tuple
}

package permissions

import (
	"context"
	"fmt"

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

// TupleFilter selects the tuples on one object, when ObjectID is set, or the tuples whose subject is one user, when
// UserID is set. Exactly one of the two is set. With UserID, a non-empty ObjectType keeps one object type. An empty
// Relation keeps every relation.
type TupleFilter struct {
	ObjectType string
	ObjectID   string
	UserID     string
	Relation   string
}

// SubjectsKey asks for the tuples in one relation of one object whose subject is UserID or is not a user.
type SubjectsKey struct {
	ObjectType string
	ObjectID   string
	Relation   string
	UserID     string
}

// Tuples stores relationship tuples for the context tenant. Every call joins the transaction the context carries, so
// a check sees that transaction's own writes, and fails as unavailable without a tenant.
type Tuples interface {
	// Subjects returns, with one statement, the tuples every key asks for.
	Subjects(ctx context.Context, keys []SubjectsKey) ([]Tuple, error)
	// Read returns the tuples the filter selects, with one statement, sorted by object, relation and subject.
	Read(ctx context.Context, filter TupleFilter) ([]Tuple, error)
	// Insert stores the tuples with one statement and keeps those already stored. It fails without a tenant even for
	// no tuples.
	Insert(ctx context.Context, tuples []Tuple) error
	// Delete removes the tuples with one statement and skips those not stored. It fails without a tenant even for no
	// tuples.
	Delete(ctx context.Context, tuples []Tuple) error
	// DeleteObject removes, with one statement, the tuples on the object and the tuples that name it as a linked
	// object or in a subject set.
	DeleteObject(ctx context.Context, objectType, objectID string) error
	// DeleteUser removes, with one statement, the tuples whose subject is the user.
	DeleteUser(ctx context.Context, userID string) error
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

func relationshipOf(tuple Tuple) (Relationship, error) {
	relationship := Relationship{
		object:   Object{objectType: ObjectType(tuple.ObjectType), id: tuple.ObjectID},
		relation: tuple.Relation,
	}
	subject := Object{objectType: ObjectType(tuple.SubjectType), id: tuple.SubjectID}
	switch {
	case tuple.SubjectType == UserSubjectType:
		user, err := uuid.Parse(tuple.SubjectID)
		if err != nil {
			return Relationship{}, fmt.Errorf("stored user subject %q: %w", tuple.SubjectID, err)
		}
		relationship.subject = userSubject(user)
	case tuple.SubjectRelation == "":
		relationship.subject = subject
	default:
		relationship.subject = subjectSet{object: subject, relation: tuple.SubjectRelation}
	}
	return relationship, nil
}

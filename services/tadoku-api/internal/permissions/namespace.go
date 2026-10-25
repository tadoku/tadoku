package permissions

import (
	"fmt"

	"github.com/google/uuid"
)

// Namespace is one object type's relations and permits. The feature that owns the object type defines it and
// builds that type's checks and relationships through it; NewModel validates it with the other namespaces.
type Namespace struct {
	objectType ObjectType
	relations  Relations
	permits    Permits
}

// MustDefine panics when the namespace refers to a relation or permit it does not define, so a wrong name stops the
// defining package from loading. NewModel checks references to other namespaces.
func MustDefine(objectType ObjectType, relations Relations, permits Permits) Namespace {
	namespace := Namespace{objectType: objectType, relations: relations, permits: permits}
	if err := namespace.validate(); err != nil {
		panic(fmt.Errorf("permissions: namespace %q: %w", objectType, err))
	}
	return namespace
}

// Check names permit on the object with objectID. A permit the namespace does not define fails the check.
func (n Namespace) Check(objectID, permit string) Check {
	return Check{object: object{objectType: n.objectType, id: objectID}, permit: permit}
}

// UserRelationship puts user in relation on the object with objectID. Writes validate it against the model.
func (n Namespace) UserRelationship(objectID, relation string, user uuid.UUID) Relationship {
	return Relationship{
		object:   object{objectType: n.objectType, id: objectID},
		relation: relation,
		subject:  userSubject(user),
	}
}

package permissions

import "fmt"

// Namespace is one object type's relations and permits. The feature that owns the object type defines it and
// builds that type's checks and relationships from its handles; NewModel validates it with the other namespaces.
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

func (n Namespace) Object(objectID string) Object {
	return Object{objectType: n.objectType, id: objectID}
}

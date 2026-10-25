package permissions

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Filter selects stored relationships by one anchor, an object or a user. Build it with OnObject or ForUser; Read
// rejects a filter without an anchor.
type Filter struct {
	anchor   anchor
	relation Relation
	in       ObjectType
	narrowed bool
}

type anchor interface {
	anchor()
}

func (Object) anchor()      {}
func (userSubject) anchor() {}

// OnObject selects the relationships whose object is object.
func OnObject(object Object) Filter {
	return Filter{anchor: object}
}

// ForUser selects the relationships whose subject is the user, not those that reach the user through a subject set.
func ForUser(userID uuid.UUID) Filter {
	return Filter{anchor: userSubject(userID)}
}

// Relation keeps the relationships in one relation, of any handle that names it.
func (f Filter) Relation(relation Relation) Filter {
	f.relation = relation
	return f
}

// In keeps a ForUser filter's relationships on objects of one namespace. Read rejects it on an OnObject filter.
func (f Filter) In(namespace Namespace) Filter {
	f.in, f.narrowed = namespace.objectType, true
	return f
}

func (m Model) tupleFilter(filter Filter) (TupleFilter, error) {
	var tuples TupleFilter
	switch anchor := filter.anchor.(type) {
	case Object:
		if err := m.validateObject(anchor); err != nil {
			return TupleFilter{}, err
		}
		if filter.narrowed {
			return TupleFilter{}, errors.New("In narrows only a ForUser filter")
		}
		tuples.ObjectType, tuples.ObjectID = string(anchor.objectType), anchor.id
	case userSubject:
		tuples.UserID = uuid.UUID(anchor).String()
		if filter.narrowed {
			if _, ok := m.namespace(filter.in); !ok {
				return TupleFilter{}, fmt.Errorf("unknown object type %q", filter.in)
			}
			tuples.ObjectType = string(filter.in)
		}
	default:
		return TupleFilter{}, errors.New("a filter needs an object or a user")
	}

	if filter.relation == nil {
		return tuples, nil
	}
	handle := filter.relation.handle()
	namespace, _ := m.namespace(handle.objectType)
	if _, ok := namespace.relations[handle.name]; !ok {
		return TupleFilter{}, fmt.Errorf("unknown relation %s#%s", handle.objectType, handle.name)
	}
	if tuples.ObjectType != "" && tuples.ObjectType != string(handle.objectType) {
		return TupleFilter{}, fmt.Errorf("%s has no relation %s#%s", tuples.ObjectType, handle.objectType, handle.name)
	}
	tuples.ObjectType, tuples.Relation = string(handle.objectType), handle.name
	return tuples, nil
}

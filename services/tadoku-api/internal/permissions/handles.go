package permissions

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// Permit is one permit of a namespace. Build it with Namespace.MustPermit in a package-level variable.
type Permit struct {
	objectType ObjectType
	name       string
}

func (p Permit) On(objectID string) Check {
	return Check{object: Object{objectType: p.objectType, id: objectID}, permit: p.name}
}

// Relation is any relation handle: a UserRelation, an ObjectRelation or a SetRelation.
type Relation interface {
	handle() relationHandle
}

type relationHandle struct {
	objectType ObjectType
	name       string
}

func (h relationHandle) handle() relationHandle { return h }

func (h relationHandle) relationship(objectID string, subject subject) Relationship {
	return Relationship{object: Object{objectType: h.objectType, id: objectID}, relation: h.name, subject: subject}
}

// UserRelation is a relation that accepts users. Build it with Namespace.MustUserRelation in a package-level
// variable.
type UserRelation struct {
	relationHandle
}

func (r UserRelation) On(objectID string, user uuid.UUID) Relationship {
	return r.relationship(objectID, userSubject(user))
}

// ObjectRelation links objects of one declared type, for Via rules to follow. Build it with
// Namespace.MustObjectRelation or MustObjectRelationOf in a package-level variable.
type ObjectRelation struct {
	relationHandle
	target ObjectType
}

func (r ObjectRelation) On(objectID, targetID string) Relationship {
	return r.relationship(objectID, Object{objectType: r.target, id: targetID})
}

// SetRelation puts the users of one declared subject set in a relation. Build it with Namespace.MustSetRelation or
// MustSetRelationOf in a package-level variable.
type SetRelation struct {
	relationHandle
	set SubjectType
}

// On names the users in the declared relation or permit of the object with setObjectID.
func (r SetRelation) On(objectID, setObjectID string) Relationship {
	set := subjectSet{object: Object{objectType: r.set.objectType, id: setObjectID}, relation: r.set.relation}
	return r.relationship(objectID, set)
}

// MustPermit panics unless the namespace defines the permit.
func (n Namespace) MustPermit(name string) Permit {
	if _, ok := n.permits[name]; !ok {
		panic(fmt.Errorf("permissions: %s has no permit %q", n.objectType, name))
	}
	return Permit{objectType: n.objectType, name: name}
}

// MustUserRelation panics unless the namespace defines the relation and it accepts users.
func (n Namespace) MustUserRelation(name string) UserRelation {
	return UserRelation{n.mustRelation(name, Users)}
}

// MustObjectRelation panics unless the relation accepts exactly one object type, which becomes the link target.
func (n Namespace) MustObjectRelation(name string) ObjectRelation {
	return n.MustObjectRelationOf(name, n.mustOnly(name, SubjectType.isObject, "object type"))
}

// MustObjectRelationOf panics unless target is an ObjectOf type the relation accepts.
func (n Namespace) MustObjectRelationOf(name string, target SubjectType) ObjectRelation {
	if !target.isObject() {
		panic(fmt.Errorf("permissions: %s#%s: %+v is not an object type", n.objectType, name, target))
	}
	return ObjectRelation{relationHandle: n.mustRelation(name, target), target: target.objectType}
}

// MustSetRelation panics unless the relation accepts exactly one subject set, whose type and relation it then uses.
func (n Namespace) MustSetRelation(name string) SetRelation {
	return n.MustSetRelationOf(name, n.mustOnly(name, SubjectType.isSet, "subject set"))
}

// MustSetRelationOf panics unless set is a SetOf subject set the relation accepts.
func (n Namespace) MustSetRelationOf(name string, set SubjectType) SetRelation {
	if !set.isSet() {
		panic(fmt.Errorf("permissions: %s#%s: %+v is not a subject set", n.objectType, name, set))
	}
	return SetRelation{relationHandle: n.mustRelation(name, set), set: set}
}

func (n Namespace) mustRelation(name string, subject SubjectType) relationHandle {
	subjects, ok := n.relations[name]
	if !ok {
		panic(fmt.Errorf("permissions: %s has no relation %q", n.objectType, name))
	}
	if !slices.Contains(subjects, subject) {
		panic(fmt.Errorf("permissions: %s#%s does not accept %+v", n.objectType, name, subject))
	}
	return relationHandle{objectType: n.objectType, name: name}
}

func (n Namespace) mustOnly(name string, kind func(SubjectType) bool, kindName string) SubjectType {
	subjects, ok := n.relations[name]
	if !ok {
		panic(fmt.Errorf("permissions: %s has no relation %q", n.objectType, name))
	}

	var matches []SubjectType
	for _, subject := range subjects {
		if kind(subject) {
			matches = append(matches, subject)
		}
	}
	if len(matches) != 1 {
		panic(fmt.Errorf("permissions: %s#%s accepts %d %ss, want one", n.objectType, name, len(matches), kindName))
	}
	return matches[0]
}

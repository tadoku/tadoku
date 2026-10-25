package permissions

import "github.com/google/uuid"

type ObjectType string

// Object is one object of a namespace. Build it with Namespace.Object.
type Object struct {
	objectType ObjectType
	id         string
}

func (o Object) Type() ObjectType { return o.objectType }

func (o Object) ID() string { return o.id }

type subject interface {
	subject()
}

type userSubject uuid.UUID

// subjectSet names the users in relation on object, where relation is a relation or a permit of object's type.
type subjectSet struct {
	object   Object
	relation string
}

func (userSubject) subject() {}
func (Object) subject()      {}
func (subjectSet) subject()  {}

// Check is a permit on one object. Build it with Permit.On.
type Check struct {
	object Object
	permit string
}

// Relationship is one stored fact whose subject is a user, a linked object or a subject set. Build it with the On
// method of a relation handle.
type Relationship struct {
	object   Object
	relation string
	subject  subject
}

func (r Relationship) Object() Object { return r.object }

func (r Relationship) Relation() string { return r.relation }

func (r Relationship) User() (uuid.UUID, bool) {
	user, ok := r.subject.(userSubject)
	return uuid.UUID(user), ok
}

func (r Relationship) Target() (Object, bool) {
	target, ok := r.subject.(Object)
	return target, ok
}

// Set returns the object and the relation or permit whose users the relationship names.
func (r Relationship) Set() (Object, string, bool) {
	set, ok := r.subject.(subjectSet)
	return set.object, set.relation, ok
}

package permissions

import "github.com/google/uuid"

type ObjectType string

type object struct {
	objectType ObjectType
	id         string
}

type subject interface {
	subject()
}

type userSubject uuid.UUID

// subjectSet names the users in relation on object. An empty relation names object itself.
type subjectSet struct {
	object   object
	relation string
}

func (userSubject) subject() {}
func (subjectSet) subject()  {}

// Check is a permit on one object. Build it with the constructors in model.go.
type Check struct {
	object object
	permit string
}

// Relationship is one stored fact. Build it with the constructors in model.go.
type Relationship struct {
	object   object
	relation string
	subject  subject
}

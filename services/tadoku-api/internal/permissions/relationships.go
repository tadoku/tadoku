package permissions

import (
	"context"

	"github.com/google/uuid"
)

// Relationships writes and reads stored relationships in the context tenant and joins the context's transaction.
// Granting a stored relationship or revoking a missing one succeeds without a change.
type Relationships struct {
	model  Model
	tuples Tuples
}

func NewRelationships(tuples Tuples, model Model) *Relationships {
	return &Relationships{model: model, tuples: tuples}
}

// Grant stores the relationships with one statement after validating all of them.
func (r *Relationships) Grant(ctx context.Context, relationships ...Relationship) error {
	tuples, err := r.tuplesOf(relationships)
	if err != nil {
		return err
	}
	return r.tuples.Insert(ctx, tuples)
}

// Revoke removes the relationships with one statement after validating all of them.
func (r *Relationships) Revoke(ctx context.Context, relationships ...Relationship) error {
	tuples, err := r.tuplesOf(relationships)
	if err != nil {
		return err
	}
	return r.tuples.Delete(ctx, tuples)
}

// RevokeObject removes every relationship on the object and every relationship that names it as a linked object or
// in a subject set, with one statement. Call it in the transaction that deletes the object.
func (r *Relationships) RevokeObject(ctx context.Context, object Object) error {
	if err := r.model.validateObject(object); err != nil {
		return err
	}
	return r.tuples.DeleteObject(ctx, string(object.objectType), object.id)
}

// RevokeUser removes every relationship whose subject is the user, with one statement.
func (r *Relationships) RevokeUser(ctx context.Context, userID uuid.UUID) error {
	return r.tuples.DeleteUser(ctx, userID.String())
}

func (r *Relationships) tuplesOf(relationships []Relationship) ([]Tuple, error) {
	tuples := make([]Tuple, len(relationships))
	for i, relationship := range relationships {
		if err := r.model.validateRelationship(relationship); err != nil {
			return nil, err
		}
		tuples[i] = tupleOf(relationship)
	}
	return tuples, nil
}

// Stored reports which of the candidates are stored, directly and without evaluating permits, in one query.
func (r *Relationships) Stored(ctx context.Context, candidates []Relationship) (map[Relationship]bool, error) {
	for _, candidate := range candidates {
		if err := r.model.validateRelationship(candidate); err != nil {
			return nil, err
		}
	}
	return storedRelationships(ctx, r.tuples, candidates)
}

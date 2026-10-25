package permissions

import "context"

// Relationships writes and reads stored relationships in the context tenant and joins the context's transaction.
// Granting a stored relationship or revoking a missing one succeeds without a change.
type Relationships struct {
	model  Model
	tuples Tuples
}

func NewRelationships(tuples Tuples, model Model) *Relationships {
	return &Relationships{model: model, tuples: tuples}
}

func (r *Relationships) Grant(ctx context.Context, relationship Relationship) error {
	if err := r.model.validateRelationship(relationship); err != nil {
		return err
	}
	return r.tuples.Insert(ctx, tupleOf(relationship))
}

func (r *Relationships) Revoke(ctx context.Context, relationship Relationship) error {
	if err := r.model.validateRelationship(relationship); err != nil {
		return err
	}
	return r.tuples.Delete(ctx, tupleOf(relationship))
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

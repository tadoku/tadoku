package permissions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Relationships writes and reads stored relationships in the context tenant and joins the context's transaction.
// Granting a stored relationship or revoking a missing one succeeds without a change.
type Relationships struct {
	model Model
	store store
}

func NewRelationships(db *pgxpool.Pool, model Model) *Relationships {
	return &Relationships{model: model, store: newStore(db)}
}

func (r *Relationships) Grant(ctx context.Context, relationship Relationship) error {
	if err := r.model.validateRelationship(relationship); err != nil {
		return err
	}
	return r.store.insert(ctx, relationship)
}

func (r *Relationships) Revoke(ctx context.Context, relationship Relationship) error {
	if err := r.model.validateRelationship(relationship); err != nil {
		return err
	}
	return r.store.delete(ctx, relationship)
}

// Stored reports which of the candidates are stored, directly and without evaluating permits, in one query.
func (r *Relationships) Stored(ctx context.Context, candidates []Relationship) (map[Relationship]bool, error) {
	for _, candidate := range candidates {
		if err := r.model.validateRelationship(candidate); err != nil {
			return nil, err
		}
	}
	return r.store.stored(ctx, candidates)
}

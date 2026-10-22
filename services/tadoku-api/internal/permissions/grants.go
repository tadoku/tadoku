package permissions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Grants writes relationships in the context tenant and joins the context's transaction. Granting a stored
// relationship or revoking a missing one succeeds without a change.
type Grants struct {
	model Model
	store store
}

func NewGrants(db *pgxpool.Pool) *Grants {
	return &Grants{model: model, store: store{db: db}}
}

func (g *Grants) Grant(ctx context.Context, relationship Relationship) error {
	if err := g.model.validateRelationship(relationship); err != nil {
		return err
	}
	return g.store.insert(ctx, relationship)
}

func (g *Grants) Revoke(ctx context.Context, relationship Relationship) error {
	if err := g.model.validateRelationship(relationship); err != nil {
		return err
	}
	return g.store.delete(ctx, relationship)
}
